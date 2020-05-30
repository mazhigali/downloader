package downloader

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Url             string // Url of file to download
	Path2save       string // if path2save = "" creates directory and downloads into it if FolderName != "" else downloads into current directory
	FolderName      string // if empty downloads to current dir
	Replace         bool   // if true it replaces already downloaded file
	Useragent       string
	Referer         string
	ProxyStr        string // "socks5://194.67.208.62:24530"
	EncryptFileName        // replaces filename encrypted with sha
}
type EncryptFileName struct {
	ShaFileName bool   // if shaFileName == true fileName from Url will be encoded. //
	Extension   string // if extension != "" shaFileName will have extension
}

// returns filename of downloaded file
func Download(conf *Config) (string, error) {
	if conf.Url == "" {
		return "", errors.New("Error Url is empty")
	}
	//fmt.Println("0", conf.Url)

	var dir_path string
	var err error
	tokens := strings.Split(conf.Url, "/")
	fileName := tokens[len(tokens)-1]
	tokens = strings.Split(fileName, ".")
	ext := tokens[len(tokens)-1]
	//fmt.Println("Downloading", urlFile, "to", fileName)

	switch conf.Path2save {
	case "":
		if conf.FolderName != "" {
			// creates directory and downloads into dir
			dir_path, err = dir_full_path(conf.FolderName)
			//fmt.Println("1", dir_path)
			if err != nil {
				return "", errors.New("Error can't get path")
			}
		} else {
			//download to current directory
			dir, _ := os.Getwd()
			dir_path = dir + string(os.PathSeparator)
			//fmt.Println("2", dir_path)
		}
	default: // download into path you assign
		if conf.FolderName != "" {
			dir_path = conf.Path2save + string(os.PathSeparator) + conf.FolderName + string(os.PathSeparator)
			//fmt.Println("3", dir_path)
		} else {
			dir_path = conf.Path2save + string(os.PathSeparator)
			//fmt.Println("4", dir_path)
		}
	}

	//fmt.Println("dir_path", dir_path)

	err = os.MkdirAll(dir_path, 0755)
	if err != nil {
		return "", fmt.Errorf("Error creating destination directory: %v", err)
	}

	var pathFile string
	var newFileName string

	// тут шифруем файлы
	if conf.EncryptFileName.ShaFileName == false {
		pathFile = dir_path + fileName
	} else {
		if conf.Extension == "" {
			newFileName = fmt.Sprintf("%x", sha1.Sum([]byte(fileName))) + "." + ext
		} else {
			newFileName = fmt.Sprintf("%x", sha1.Sum([]byte(fileName))) + "." + conf.Extension
		}
		pathFile = dir_path + newFileName
	}

	if conf.Replace == false { //если замена выставлена на true то заменяем файлы, если нет, то пропускаем
		//если файл существует, то выходим из функции и возвращаем имя файла, если ошибка то начинаем скачивание
		if _, err := os.Stat(pathFile); err == nil {
			//fmt.Println("file already exists")
			return fileName, nil
			//} else {
			//return "", err
		}
	}

	output, err := os.Create(pathFile)
	if err != nil {
		return "", fmt.Errorf("Error while creating: %v -  %v", fileName, err)
	}
	defer output.Close()

	//creating the proxyURL
	proxyURL, err := url.Parse(conf.ProxyStr)
	if err != nil {
		return "", fmt.Errorf("Error while reading proxy string: %v", err)
	}
	//adding the proxy settings to the Transport object
	transport := &http.Transport{}
	if conf.ProxyStr != "" {
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			//TLSClientConfig: &tls.Config{},
			//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	clientWithCheck := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: transport,
	}

	request, err := http.NewRequest("GET", conf.Url, nil)
	if err != nil {
		return "", fmt.Errorf("Error while request: %v", err)
	}

	if conf.Useragent != "" {
		request.Header.Add("User-Agent", conf.Useragent)
	}
	if conf.Referer != "" {
		request.Header.Add("Referer", conf.Referer)
	}
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	request.Header.Add("Accept-Language", "ru,en-US;q=0.7,en;q=0.3")

	response, err := clientWithCheck.Do(request)
	if err != nil {
		fmt.Printf("Error while downloading %v\nDeleting file: %v\nErr: %v", conf.Url, pathFile, err)
		err = os.Remove(pathFile)
		if err != nil {
			return "", fmt.Errorf("Error while remove file: %v", err)
		}
		return "", fmt.Errorf("Error while getting response.\nLocal file deleted: %v - %v", conf.Url, err)
	}
	defer response.Body.Close()

	if strings.Contains(response.Header.Get("Content-Type"), "text/html") == true {
		err = os.Remove(pathFile)
		if err != nil {
			return "", fmt.Errorf("Error while remove file: %v", err)
		}
		return "", fmt.Errorf("Error: can't download: GOT HTML %v", conf.Url)
	}
	if response.ContentLength <= 0 {
		err = os.Remove(pathFile)
		if err != nil {
			return "", fmt.Errorf("Error while remove file: %v", err)
		}
		return "", fmt.Errorf("Error: invalid content length %v", conf.Url)
	}

	//записываем ответ от сервера в файл
	size, errCopy := io.Copy(output, response.Body)
	if err != nil {
		//fmt.Println("Error while io Copy", conf.Url, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", fmt.Errorf("Error deleting File: %v", err)
		}
		return "", fmt.Errorf("Error while io Copy: %v - %v", conf.Url, errCopy)
	}

	if size == 0 {
		//fmt.Println("Error while downloading null size File", conf.Url, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", fmt.Errorf("Error deleting File: %v - %v", pathFile, err)
		}
		return "", fmt.Errorf("Zero size file Downloaded: %v \nFile deleted %v", conf.Url, pathFile)
	}

	//fmt.Println("HEADER response:", response.Header)
	//fmt.Println("HEADER request:", request.Header)

	//fmt.Println(size, "bytes downloaded.")
	return fileName, nil
}

//get path 2 current dir
func dir_full_path(parsSite string) (string, error) {
	path, err := filepath.Abs(parsSite)

	if err != nil {
		fmt.Println("sdfsdfsdfsdf")
		return "", err
	}

	//t := time.Now()

	s := path +
		//string(os.PathSeparator) +
		//strconv.Itoa(t.Day()) +
		//"_" +
		//strconv.Itoa(int(t.Month())) +
		//"_" +
		//strconv.Itoa(t.Year()) +
		string(os.PathSeparator)

	return s, nil
}

func SplitAndGetName(stroka string, razdelitel string, numIzMassiva int) string {
	tokens := strings.Split(stroka, razdelitel)
	nameProduct := tokens[len(tokens)-numIzMassiva]
	return nameProduct
}
