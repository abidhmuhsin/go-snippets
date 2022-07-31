//https://stackoverflow.com/questions/23005917/parallel-zip-compression-in-go
/*
Run by command.
go build async-zipping-channel.go && ./async-zipping-channel async-zipping-channel.go async-zipping-channel file 2 file 3 ...

A helper method ZipWriter receives a channel of os.Files and spawns a goroutine to add all files from the channel into a zip file.
A waitgroup is returned from the helper method which can be used to await zipping completion.
Main create the channel and invokes the ZipWriter, then waits for zipping completion.
Another go routine in main will keep on sending files passed as arguments list to the created channel.

-->  sending to channel -- can be async and logs appear one after another irrespective of zipping
<--  received in channel for zipping  -- can occur only after corresponding send is logged.
	 cannot appear again unless zipping completed(<> logged) for in progress file.
<>   zipped in channel. -- can appear only after corresponding file is received in channel.
*/
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"sync"
)

func main() {
	// Create a channel of files and pass as argument to ZipWriter which will listen to passed channel for zipping files.
	// If the capacity is zero or absent, the channel is unbuffered and communication succeeds only when both a sender and receiver are ready.
	// If the channel is unbuffered, the sender blocks until the receiver has received the value
	files := make(chan *os.File)
	wg_AwaitZipCompleteRef := ZipWriter(files) // store the returned waitgroup for awaiting zip processing completed from inside fn go-routine.

	// Send all files to the zip writer.

	// Create a new waitgroup and add the number of files (length of arguments) to it.
	// Call Done for each file and wg will be complete once all files are read.
	var wg_SendFile sync.WaitGroup
	wg_SendFile.Add(len(os.Args) - 1)
	// Read all filenames from command line args
	for i, name := range os.Args {
		if i == 0 {
			// no arguments
			continue
		}
		// Read/Open each file in parallel:
		go func(name string) {
			fmt.Printf("-->[%s] ::Sending file to channel. \n", name)
			defer wg_SendFile.Done()
			f, err := os.Open(name)
			if err != nil {
				panic(err)
			}
			// send each file to channel
			files <- f
		}(name)
	}

	wg_SendFile.Wait()
	// Once we're done sending the files, we can close the channel.
	close(files)
	// This will cause ZipWriter to break out of the loop (**REF#1 since files channel is no more), close the file,
	// and unblock the next mutex:
	wg_AwaitZipCompleteRef.Wait()
}

/*
Below function accepts a channel of files which will be listned for receiving new files to zip.
It creates and returns a new waitgroup reference that can be used by the caller to wait until the zipping operations are completed.
The wg will be done once the files channel is fully processed in the goroutine
*/

func ZipWriter(filesChannel chan *os.File) *sync.WaitGroup {
	zippedOutFile, err := os.Create("out.zip")
	if err != nil {
		panic(err)
	}
	var wgRef sync.WaitGroup
	wgRef.Add(1)
	zw := zip.NewWriter(zippedOutFile)
	// Spawn a go-routine that listens for files on the filesChannel
	go func() {
		// Note the order of files zipped (LIFO):
		defer wgRef.Done()          // 2. signal (the fn caller) that we're done. Will be run once we break out of below for loop and exit this fn.
		defer zippedOutFile.Close() // 1. close the file
		var err error
		var fw io.Writer
		for f := range filesChannel {
			//REF#1 Loop until channel filesChannel is closed.
			fmt.Printf("<--[%s] ::Received file in zipping channel.\n", f.Name())
			if fw, err = zw.Create(f.Name()); err != nil {
				panic(err)
			}
			io.Copy(fw, f)
			if err = f.Close(); err != nil {
				panic(err)
			}
			fmt.Printf("<> [%s] ::Zipping complete for file.\n", f.Name())
		}
		// The zip writer must be closed *before* zippedOutFile.Close() is called (via defer)!
		if err = zw.Close(); err != nil {
			panic(err)
		}
	}()
	// Caller can use this wg to await completion of zipping goroutine
	return &wgRef
}
