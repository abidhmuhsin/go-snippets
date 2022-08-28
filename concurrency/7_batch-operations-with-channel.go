package main

import (
	"fmt"
	"time"
)

type Job struct {
	Id int
}

const MaxBatchSize = 50

func main() {

	jobChannel := make(chan Job, 1000) // Using the buffered job channel provides a form of rate limiting out of the box.

	go worker(jobChannel)
	// time.Sleep(1 * time.Millisecond)
	for i := 0; i < 5000; i++ {
		/*In fact, if this was an operation you'd do based on HTTP requests you have no control over the speed
		as opposed to a list of jobs which you have control over, all you'd have to do is
		check if the length of the channel is the same as the capacity of the channel,
		and if the length has reached the capacity, return 429: Too many requests.
		*/
		// if len(jobChannel) == cap(jobChannel) {
		// 	// reached job channel capacity, return 429 here for capping http requests
		// 	fmt.Printf("reached job channel capacity: %d\n", cap(jobChannel))
		// 	break // return
		// }
		jobChannel <- Job{Id: i + 1}
	}
	// return
	// wait for channel to be empty
	for len(jobChannel) != 0 {
		fmt.Printf("Waiting for channel to be empty\n")
		time.Sleep(100 * time.Millisecond)
	}
}

func worker(jobChannel <-chan Job) {
	var jobsInBatch []Job
	for {
		if len(jobChannel) > 0 && len(jobsInBatch) < MaxBatchSize {
			jobsInBatch = append(jobsInBatch, <-jobChannel)
			continue
		}
		isBatchFull := len(jobsInBatch) == MaxBatchSize
		isLastBatch := len(jobChannel) == 0 && len(jobsInBatch) > 0
		if isBatchFull || isLastBatch {
			// process the job
			str := fmt.Sprintf("processing batch of %d jobs[%d-%d]. Remaining jobs in channel buffer %d.\n",
				len(jobsInBatch),                   // total jobs in batch
				jobsInBatch[0].Id,                  // first job in batch
				jobsInBatch[len(jobsInBatch)-1].Id, // last job in batch
				len(jobChannel))                    // total jobs pending in channel buffer
			go doSometaskWithBatch(str)
			// clear the list of jobs that were just processed
			jobsInBatch = jobsInBatch[:0] // remove all
		}
		// No jobs in the channel? back off.
		if len(jobChannel) == 0 {
			fmt.Println("...No more jobs in the channel..Backing off...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func doSometaskWithBatch(str string) {
	fmt.Print(str)
}
