package main

import "time"

func main() {
	finished := make(chan bool) //  creates a channel that expects booleans as data type
	go func() {
		time.Sleep(time.Second * 3)
		println("World")
		finished <- true //finished <- true sends the data true to the channel finished
	}()
	println("Hello")
	<-finished //waits for data to be sent to the channel finished.l
}
