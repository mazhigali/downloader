// Package main provides ...
package main

import (
	"fmt"
	"log"

	"github.com/mazhigali/downloader"
)

func main() {
	//filename, err := downloader.DownloadFromUrl("https://www.archlinux.org/static/archnavbar/archlogo.a2d0ef2df27d.png", "", "", true, false, "")
	filename, err := downloader.Download(downloader.Config{
		Url:        "https://http2.golang.org/file/gopher.png",
		Path2save:  "",
		FolderName: "",
		Replace:    true,
		Useragent:  "Mozilla/5.0 (X11; Linux x86_64; rv:76.0) Gecko/20100101 Firefox/76.0",
		ProxyStr:   "",
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
