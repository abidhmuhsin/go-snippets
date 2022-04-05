package main

/*
Uploads via formfile:myFile
Validates filesize < 1mb, MIME filetype jpg/png
Reads into memory fully, then writes(copies) to ./uploads folder

*/
import (
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"time"
)

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
						<input type="file" name="myFile" />
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

	//--------------------Set the maximum file size--------------------
	/*
		if r.ContentLength > MAX_UPLOAD_SIZE {
			//Content-Length header can be modified on the client to be any value regardless of the actual file size
			http.Error(w, "The uploaded image is too big. Please use an image less than 1MB in size", http.StatusBadRequest)
			return
		}
	*/
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
		return
	}

	//--------------------Save the uploaded file--------------------

	// The argument to FormFile must match the name attribute
	// of the file input on the frontend
	file, fileHeader, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	//--START:Optionally Restrict the type of the uploaded file------------
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The DetectContentType() method is provided by the http package for the purpose of detecting the content type of the given data.
	// It considers (at most) the first 512 bytes of data to determine the MIME type.
	// This is why we read the first 512 bytes of the file to an empty buffer before passing it to the DetectContentType() method.
	// If the resulting filetype is neither a JPEG or PNG, an error is returned.
	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		http.Error(w, "The provided file format is not allowed. Please upload a JPEG or PNG image", http.StatusBadRequest)
		return
	}

	// When we read the first 512 bytes of the uploaded file in order to determine the content type,
	// the underlying file stream pointer moves forward by 512 bytes. When io.Copy() is called later,
	// it continues reading from that position resulting in a corrupted image file.
	// The file.Seek() method is used to return the pointer back to the start of the file so that io.Copy() starts from the beginning.
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// --END:Restrict type------------------

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(fmt.Sprintf("./uploads/%d%s", time.Now().UnixNano(), filepath.Ext(fileHeader.Filename)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Upload successful")
}
