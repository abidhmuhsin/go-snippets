package main

// Basic go file upload handler

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
	http.HandleFunc("/upload", uploadFileHandler)
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
func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")
	/*
		ParseMultipartForm will parse every single of part of request,
		if you send 100 files in a “chunked” encoding POST to the above endpoint,
		it will parse them all. Notice that 32Mb is the bytes allocated to the request body to store in memory,
		not the limit of request body, when full (says 33Mb), it will write to temporary directory.
		use defer ParseMultipartForm.RemoveAll() to remove any temp files created.
	*/
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	// ?? r.ParseMultipartForm(10 << 20)  wont stop processing at set limit 10MB, for stopping automatically when crossing 10MB use
	// ?? r.Body = http.MaxBytesReader(w, r.Body, 10<<20) just line above
	r.ParseMultipartForm(10 << 20) // Parse our multipart form, 10 << 20 specifies a maximum of 10 MB live in RAM.
	// FormFile returns the first file(full file will be streamed to /tmp/multipart-xyz) for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, fileHeader, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
	fmt.Printf("File Size: %+v\n", fileHeader.Size)
	fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a temporary file in ./uploads that follows a particular naming pattern
	// tempFile, err := ioutil.TempFile("./uploads", "upload-*.TMP")

	// OR Create a new file in the uploads subdirectory
	uniqueFileName := fmt.Sprintf("./uploads/%s_%d%s", filepath.Base(fileHeader.Filename), time.Now().UnixNano(), filepath.Ext(fileHeader.Filename))
	tempFile, err := os.Create(uniqueFileName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	/***
	//IN MEMORY READ, LARGE FILES WILL BREAK THE APPLICATION
	// read all of the contents of our uploaded file into a
	// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println(err)
		}
		// write this byte array to our temporary file
		tempFile.Write(fileBytes)
	***/
	/*** STREAMING READ ****/
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File \n%s\n", tempFile.Name())
}
