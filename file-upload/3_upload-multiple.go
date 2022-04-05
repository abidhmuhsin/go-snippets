package main

/*
Uploads multiple files in one request via formfile:myMultipleFiles
Validates filesize < 1mb, MIME filetype jpg/png
Reads everything into memory fully, then writes(copies) to ./uploads folder
*/
import (
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

var s3Client *minio.Client

func main() {
	setupRoutes()
}

func setupRoutes() {
	http.HandleFunc("/", landingPageHandler)
	http.HandleFunc("/upload", uploadHandler)
	err := http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Landing page Hit")
	const template = `
					<!DOCTYPE html>
					<html lang="en">
					<head>
						<meta charset="UTF-8" />
						<meta name="viewport" content="width=device-width, initial-scale=1.0" />
						<meta http-equiv="X-UA-Compatible" content="ie=edge" />
						<title>Document</title>
					</head>
					<body>
						<form
						enctype="multipart/form-data"
						action="http://localhost:6060/upload"
						method="post"
						>
						<input type="file" name="myMultipleFiles" multiple />
						<input type="submit" value="upload" />
						</form>
					</body>
					</html>
					`
	fmt.Fprintf(w, template)
}

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 32 MB is the default used by FormFile()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["myMultipleFiles"] // use name="myMultipleFiles" and add `multiple` attribute to input tag

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MAX_UPLOAD_SIZE {
			http.Error(w, fmt.Sprintf("The uploaded image is too big: \n\n%s.\n\nPlease use an image less than 1MB in size", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filetype := http.DetectContentType(buff)
		if filetype != "image/jpeg" && filetype != "image/png" {
			http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
			return
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.MkdirAll("./uploads", os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		f, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	fmt.Fprintf(w, "Upload successful")
}
