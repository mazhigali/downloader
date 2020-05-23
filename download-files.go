package downloadFiles

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Url       string
	Path2save string // if path2save = "_" it downloads it into current directory  if "" creates directory and downloads into dir
	ParsSite  string //adds folderName 2 downloadPath
	Zamena    bool   // if true it replaces already downloaded file
	Useragent string
	ProxyStr  string
	EncryptFileName
}
type EncryptFileName struct {
	ShaFileName bool   // if shaFileName == true fileName from Url will be encoded. //
	Extension   string // if extension != "" shaFileName will have extension
}

func Download(conf Config) (string, error) {
	var dir_path string
	var err error
	tokens := strings.Split(conf.Url, "/")
	fileName := tokens[len(tokens)-1]
	tokens = strings.Split(fileName, ".")
	ext := tokens[len(tokens)-1]
	//fmt.Println("Downloading", urlFile, "to", fileName)

	switch conf.Path2save {
	case "_": //download to current directory
		dir, _ := os.Getwd()
		dir_path = dir + string(os.PathSeparator)

	case "": // creates directory and downloads into dir
		dir_path, err = dir_full_path(conf.ParsSite)
		if err != nil {
			fmt.Println("Error while getting path")
			return "", err
		}
	default: // download into path you assign
		if conf.ParsSite != "" {
			dir_path = conf.Path2save + string(os.PathSeparator) + conf.ParsSite + string(os.PathSeparator)
		} else {
			dir_path = conf.Path2save + string(os.PathSeparator)
		}
	}

	err = os.MkdirAll(dir_path, 0755)
	if err != nil {
		fmt.Errorf("error creating destination directory: %v", err)
		return "", err
	}

	var pathFile string
	var newFileName string

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

	if conf.Zamena == false { //если замена выставлена на true то заменяем файлы, если нет, то пропускаем
		//если файл существует, то выходим из функции и возвращаем имя файла, если ошибка болт возвращаем
		if _, err := os.Stat(pathFile); err == nil {
			//fmt.Println("file already exists")
			return fileName, nil
			//} else {
			//return "", err
		}
	}

	output, err := os.Create(pathFile)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return "", err
	}
	defer output.Close()

	//creating the proxyURL
	proxyURL, err := url.Parse(conf.ProxyStr)
	if err != nil {
		fmt.Println("Error while reading proxy string", err)
		return "", err
	}
	//adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{},
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: transport,
	}

	req, err := http.NewRequest("GET", conf.Url, nil)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}

	req.Header.Add("User-Agent", conf.Useragent)
	req.Header.Add("Accept", "*/*")

	response, err := check.Do(req)
	//response, err := check.Get(urlFile)
	//response, err := http.Get(urlFile)
	if err != nil {
		fmt.Println("Error while downloading", conf.Url, "-", err)
		err = os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error while remove file")
		}
		return "", errors.New("Error while request")
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Type") == "text/html" {
		return "", errors.New("Error can't download: GOT HTML")
	}

	//записываем ответ от сервера в файл
	size, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while io Copy", conf.Url, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", err
	}

	if size == 0 {
		fmt.Println("Error while downloading null size File", conf.Url, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", errors.New("Zero size file Downloaded")
	}

	//fmt.Println(size, "bytes downloaded.")
	return fileName, nil
}

//input URL to download and path2Save without \/ or if path2save empty it creates directory in current dir > output filename
// if path2save = "_" it downloads it into current directory
// if shaFileName == true fileName from Url will be encoded. //
// if extension != "" shaFileName will have extension
func DownloadFromUrl(url string, path2save string, parsSite string, zamena bool, shaFileName bool, extension string) (string, error) {
	var dir_path string
	var err error
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	tokens = strings.Split(fileName, ".")
	ext := tokens[len(tokens)-1]
	//fmt.Println("Downloading", url, "to", fileName)

	switch path2save {
	case "_": //download to current directory
		dir, _ := os.Getwd()
		dir_path = dir + string(os.PathSeparator)

	case "": // creates directory and downloads into dir
		dir_path, err = dir_full_path(parsSite)
		if err != nil {
			fmt.Println("Error while getting path")
			return "", err
		}
	default: // download into path you assign
		if parsSite != "" {
			dir_path = path2save + string(os.PathSeparator) + parsSite + string(os.PathSeparator)
		} else {
			dir_path = path2save + string(os.PathSeparator)
		}
	}

	err = os.MkdirAll(dir_path, 0755)
	if err != nil {
		fmt.Errorf("error creating destination directory: %v", err)
		return "", err
	}

	var pathFile string
	var newFileName string

	if shaFileName == false {
		pathFile = dir_path + fileName
	} else {
		if extension == "" {
			newFileName = fmt.Sprintf("%x", sha1.Sum([]byte(fileName))) + "." + ext
		} else {
			newFileName = fmt.Sprintf("%x", sha1.Sum([]byte(fileName))) + "." + extension
		}
		pathFile = dir_path + newFileName
	}

	if zamena == false { //если замена выставлена на true то заменяем файлы, если нет, то пропускаем
		//если файл существует, то выходим из функции и возвращаем имя файла, если ошибка болт возвращаем
		if _, err := os.Stat(pathFile); err == nil {
			//fmt.Println("file already exists")
			if newFileName == "" {
				return fileName, nil
			} else {
				return newFileName, nil
			}
			//} else {
			//return "", err
		}
	}

	output, err := os.Create(pathFile)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return "", err
	}
	defer output.Close()

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (iPad; CPU OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko)Version/9.0 Mobile/13B143 Safari/601.1")
	req.Header.Add("Accept", "*/*")

	response, err := check.Do(req)
	//response, err := check.Get(url)
	//response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		err = os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error while request")
		}
		return "", errors.New("Error while request")
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Type") == "text/html" {
		return "", errors.New("Error can't download: GOT HTML")
	}

	//записываем ответ от сервера в файл
	size, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", err
	}

	if size == 0 {
		fmt.Println("Error while downloading null size File", url, "-", err)

		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", errors.New("Zero size file Downloaded")

	}

	//fmt.Println(size, "bytes downloaded.")
	if newFileName == "" {
		return fileName, nil
	} else {
		return newFileName, nil
	}
}

func DownloadFromUrlProxyAndUa(urlFile string, path2save string, parsSite string, zamena bool, proxyStr string, userAgent string) (string, error) {
	var dir_path string
	var err error
	tokens := strings.Split(urlFile, "/")
	fileName := tokens[len(tokens)-1]
	//fmt.Println("Downloading", urlFile, "to", fileName)

	switch path2save {
	case "_": //download to current directory
		dir, _ := os.Getwd()
		dir_path = dir + string(os.PathSeparator)

	case "": // creates directory and downloads into dir
		dir_path, err = dir_full_path(parsSite)
		if err != nil {
			fmt.Println("Error while getting path")
			return "", err
		}
	default: // download into path you assign
		if parsSite != "" {
			dir_path = path2save + string(os.PathSeparator) + parsSite + string(os.PathSeparator)
		} else {
			dir_path = path2save + string(os.PathSeparator)
		}
	}

	err = os.MkdirAll(dir_path, 0755)
	if err != nil {
		fmt.Errorf("error creating destination directory: %v", err)
		return "", err
	}

	pathFile := dir_path + fileName

	if zamena == false { //если замена выставлена на true то заменяем файлы, если нет, то пропускаем
		//если файл существует, то выходим из функции и возвращаем имя файла, если ошибка болт возвращаем
		if _, err := os.Stat(pathFile); err == nil {
			//fmt.Println("file already exists")
			return fileName, nil
			//} else {
			//return "", err
		}
	}

	output, err := os.Create(pathFile)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return "", err
	}
	defer output.Close()

	//creating the proxyURL
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		fmt.Println("Error while reading proxy string", err)
		return "", err
	}
	//adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{},
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: transport,
	}

	req, err := http.NewRequest("GET", urlFile, nil)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}

	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Accept", "*/*")

	response, err := check.Do(req)
	//response, err := check.Get(urlFile)
	//response, err := http.Get(urlFile)
	if err != nil {
		fmt.Println("Error while downloading", urlFile, "-", err)
		err = os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error while remove file")
		}
		return "", errors.New("Error while request")
	}
	defer response.Body.Close()

	if response.Header.Get("Content-Type") == "text/html" {
		return "", errors.New("Error can't download: GOT HTML")
	}

	//записываем ответ от сервера в файл
	size, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while io Copy", urlFile, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", err
	}

	if size == 0 {
		fmt.Println("Error while downloading null size File", urlFile, "-", err)
		err := os.Remove(pathFile)
		if err != nil {
			return "", errors.New("Error deleting File")
		}
		return "", errors.New("Zero size file Downloaded")
	}

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
