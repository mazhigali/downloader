// Package main provides ...
package main

import (
	"fmt"
	"log"

	"github.com/mazhigali/downloader"
)

func main() {
	//filename, err := downloader.DownloadFromUrl("https://www.archlinux.org/static/archnavbar/archlogo.a2d0ef2df27d.png", "", "", true, false, "")
	filename, err := downloader.Download(&downloader.Config{
		Url: "ftp://ftp_drive_d_r:zP3CxVm4O8kg5UWkG5D@cloud.datastrg.ru:21/649c325f-9b96-4c59-b277-afa9bcc0cb42___v8_3730_34ec.jpeg",
		//Url: "https://my.zadarma.com/images/logo2.png",
		//Url:        "https://http2.golang.org/reqinfo",
		//Url:        "https://http2.golang.org/redirect",
		Path2save:  "/tmp",
		FolderName: "",
		Replace:    true,
		Useragent:  "Mozilla/5.0 (X11; Linux x86_64; rv:76.0) Gecko/20100101 Firefox/76.0",
		//ProxyStr:   "socks5://194.67.208.62:24530",
		EncryptFileName: downloader.EncryptFileName{
			ShaFileName: false,
			Extension:   "",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(filename)
}
