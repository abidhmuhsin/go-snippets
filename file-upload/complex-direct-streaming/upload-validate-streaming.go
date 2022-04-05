package main

// ref: https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
/*
To protect upload api from being hogged with extremely large file uploaded.

Handles file upload without fully reading the uploaded file into memory.
File upload code to read only first 512bytes of a file from a multipart form.
Then verify the filetype and then use a limitreader to cut off upload stream at size limit.
*/
import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// function body of a http.HandlerFunc
	// limit the POST body size to 32.5Mb and throw error if client attempt to send more than that.
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024)

	/*  Use MultipartReader to stream instead of ParseMultipartForm

	ParseMultipartForm will parse every single of part of request,
	if you send 100 files in a “chunked” encoding POST to the above endpoint, it will parse them all.
	Say when using r.ParseMultipartForm(32 << 20),
	Notice that 32Mb is the bytes allocated to the request body to store in memory,
	not the limit of request body, when full (say 33Mb), it will write to temporary directory.

	In the below code, the main things we do are
	-- Only fetch first 2 parts in POST body, handler will expect text_field then file_field
	-- Assert file type using bufio.Reader wrapping Part reader (peeking first 512 bytes)
	-- Limit file size with io.LimitReader (with 1 byte offset to see if part reader still has some data left)
	*/

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// parse text field
	text := make([]byte, 512)
	p, err := mr.NextPart()
	// one more field to parse, EOF is considered as failure here
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.FormName() != "text_field" {
		http.Error(w, "text_field is expected", http.StatusBadRequest)
		return
	}
	_, err = p.Read(text)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// parse file field
	p, err = mr.NextPart()
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.FormName() != "file_field" {
		http.Error(w, "file_field is expected", http.StatusBadRequest)
		return
	}
	buf := bufio.NewReader(p)
	sniff, _ := buf.Peek(512) //Peek returns the next n bytes without advancing the reader
	//DetectContentType uses the first 512 bytes of data to determine the MIME type.
	contentType := http.DetectContentType(sniff)
	if contentType != "application/zip" {
		http.Error(w, "file type not allowed", http.StatusBadRequest)
		return
	}
	f, err := ioutil.TempFile("", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	var maxSize int64 = 32 << 20 // 32MB
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))
	written, err := io.Copy(f, lmt)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if written > maxSize {
		os.Remove(f.Name())
		http.Error(w, "file size over limit", http.StatusBadRequest)
		return
	}
	// schedule for other stuffs (s3, scanning, etc.)
}
