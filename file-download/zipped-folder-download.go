package main

import (
	"archive/zip"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/minio/minio-go/v7"
)

var s3Client *minio.Client

func main() {
	setupRoutes()
}

func setupRoutes() {
	http.HandleFunc("/", downloadFileHandler)
	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}
	// go func() {
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Hit downloadFileHandler")

	downloadName := time.Now().UTC().Format("data-20060102150405")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+downloadName+".zip")
	w.WriteHeader(http.StatusOK)

	zipFile := zip.NewWriter(w)
	defer zipFile.Close()

	file1, _ := zipFile.Create(downloadName + ".xlsx") // Create file#1 in zipFile
	// file2, _ := zipFile.Create("another_file.xlsx") //?? Create file#2 in zipFile
	// excelFile.WriteTo(file1)
	return
}
