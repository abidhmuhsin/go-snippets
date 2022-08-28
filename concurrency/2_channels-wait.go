package main

import "time"

func main() {
	worldChannel := make(chan string)
	dearChannel := make(chan string)
	go func() {
		time.Sleep(time.Second * 3)
		worldChannel <- "world"
	}()
	go func() {
		time.Sleep(time.Second * 2)
		dearChannel <- "dear"
	}()
	println("Hello", <-dearChannel, <-worldChannel) //waits for all data to be sent to the channel before printing
}
