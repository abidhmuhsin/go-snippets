package main

/*
Upload handler that shows progress on server using TeeReader.
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

	// FormFile returns the first file(full file will be streamed to /tmp/multipart-xyz)  for the given key `myFile`
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

	/*** STREAMING READ from source to dest, source is file already loaded in ram ****/
	// _, err = io.Copy(f, file) ------- replace with -- io.Copy(f, io.TeeReader(file, pr)) -- for progress
	//TeeReader returns a Reader that writes to w what it reads from r. All reads from r performed through it are matched with corresponding writes to w.
	//There is no internal buffering - the write must complete before the read completes. Any error encountered while writing is reported as a read error.
	pr := &Progress{
		TotalSize: fileHeader.Size,
	}

	_, err = io.Copy(tempFile, io.TeeReader(file, pr))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File \n%s\n", tempFile.Name())
}

//--START:Progress Helper Struct---------------

// Progress is used to track the progress of a file upload.
// It implements the io.Writer interface so it can be passed
// to an io.TeeReader()
type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface.
// Instead of writing somewhere, it simply aggregates
// the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil
	pr.BytesRead += int64(n)
	pr.Print()
	return
}

// Print displays the current progress of the file upload
// each time Write is called
func (pr *Progress) Print() {

	percentRead := (pr.BytesRead * 100) / pr.TotalSize
	fmt.Printf("File upload in progress:[%d%%]  %d/%d Bytes\n", percentRead, pr.BytesRead, pr.TotalSize)
	//Use the special verb %%, which consumes no argument, to write a literal percent sign:
	//fmt.Printf("%d %%", 50) // prints "50 %"

	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")
		return
	}
}

//--END:Progress Helper Struct-------------
