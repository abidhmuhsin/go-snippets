// Reference: https://www.reddit.com/r/golang/comments/apf6l5/multiple_files_upload_using_gos_standard_library/
//
// Original author: https://www.reddit.com/user/teizz
//
// handles multiple files being uploaded in one request.
// reads them in blocks of 4K
// writes them to a temporary file in $TMPDIR
// calculates and logs the SHA256 sum
// proceeds to remove the temporary file through a defer statement
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	indexPage = `<html>
	<body>
		<form enctype="multipart/form-data" action="http://localhost:6060/upload" method="post">
			<input type="file" name="files" multiple />
			<input type="submit" value="upload" />
		</form>
	</body>
</html>`
)

func doSomethingWithFile(f *os.File) {
	// Below snip calculates SHA sum.
	if n, err := f.Seek(0, 0); err != nil || n != 0 {
		log.Printf("unable to seek to beginning of file '%s'", f.Name())
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Printf("unable to hash '%s': %s", f.Name(), err.Error())
	}
	log.Printf("SHA256 sum of '%s': %x", f.Name(), h.Sum(nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(indexPage))
}

func readParts(w http.ResponseWriter, r *http.Request) {
	// define some variables used throughout the function
	// n: for keeping track of bytes read and written
	// err: for storing errors that need checking
	var n int
	var err error

	// define pointers for the multipart reader and its parts
	var mr *multipart.Reader
	var part *multipart.Part

	log.Println("File Upload Endpoint Hit")
	/*
		ParseMultipartForm will parse every single of part of request,
		if you send 100 files in a “chunked” encoding POST to the above endpoint,
		it will parse them all. Notice that 32Mb is the bytes allocated to the request body to store in memory,
		not the limit of request body, when full (says 33Mb), it will write to temporary directory.

		MultipartReader returns a MIME multipart reader if this is a multipart/form-data or a multipart/mixed POST request,
		else returns nil and an error.
		Use this function instead of ParseMultipartForm to process the request body as a stream.

		https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad

	*/
	if mr, err = r.MultipartReader(); err != nil {
		log.Printf("Hit error while opening multipart reader: %s", err.Error())
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error occured during upload")
		return
	}

	// buffer to be used for reading bytes from files
	chunk := make([]byte, 4096) //multipart package source code, the peek buffer size is set as constant to 4096

	// continue looping through all parts, *multipart.Reader.NextPart() will
	// return an End of File when all parts have been read.
	for {
		// variables used in this loop only
		// tempfile: filehandler for the temporary file
		// filesize: how many bytes where written to the tempfile
		// uploaded: boolean to flip when the end of a part is reached
		var tempfile *os.File
		var filesize int
		var uploaded bool

		if part, err = mr.NextPart(); err != nil {
			if err != io.EOF {
				log.Printf("Hit error while fetching next part: %s", err.Error())
				w.WriteHeader(500)
				fmt.Fprintf(w, "Error occured during upload")
			} else {
				log.Printf("Hit last part of multipart upload")
				w.WriteHeader(200)
				fmt.Fprintf(w, "Upload completed")
			}
			return
		}
		// at this point the filename and the mimetype is known
		log.Printf("Uploaded filename: %s", part.FileName())
		log.Printf("Uploaded mimetype: %s", part.Header)

		tempfile, err = ioutil.TempFile(os.TempDir(), "example-upload-*.tmp")
		if err != nil {
			log.Printf("Hit error while creating temp file: %s", err.Error())
			w.WriteHeader(500)
			fmt.Fprintf(w, "Error occured during upload")
			return
		}
		defer tempfile.Close()

		// defer the removal of the tempfile as well, something can be done
		// with it before the function is over (as long as you have the filehandle)
		defer os.Remove(tempfile.Name())

		// here the temporary filename is known
		log.Printf("Temporary filename: %s\n", tempfile.Name())

		// continue reading until the whole file is upload or an error is reached
		for !uploaded {
			if n, err = part.Read(chunk); err != nil {
				if err != io.EOF {
					log.Printf("Hit error while reading chunk: %s", err.Error())
					w.WriteHeader(500)
					fmt.Fprintf(w, "Error occured during upload")
					return
				}
				uploaded = true // ie err==io.EOF, n will have last part now.
			}
			// ?? May try bufio for larger buffer bufio.NewReader(p)

			// Do something with each part like writing to tempfile stream or another outputstream
			if n, err = tempfile.Write(chunk[:n]); err != nil {
				log.Printf("Hit error while writing chunk: %s", err.Error())
				w.WriteHeader(500)
				fmt.Fprintf(w, "Error occured during upload")
				return
			}
			filesize += n
			log.Printf("Read: %d bytes [%.2fMB]", filesize, (float64(filesize) / (1024 * 1024)))
		}
		log.Printf("Uploaded filesize: %d bytes", filesize)

		// once uploaded something can be done with the file, the last defer
		// statement will remove the file after the function returns so any
		// errors during upload won't hit this, but at least the tempfile is
		// cleaned up
		doSomethingWithFile(tempfile)
	}
}

func main() {
	log.Println("Gopher upload service started")
	http.HandleFunc("/upload", readParts)
	http.HandleFunc("/", serveIndex)
	log.Fatalf("Exited: %s", http.ListenAndServe(":6060", nil))
}
