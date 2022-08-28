package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	Id int
}

type JobResult struct {
	Output string
}

const NumberOfWorkers = 10

func main() {
	start := time.Now()

	// Create fake jobs for testing purposes
	var jobs []Job
	for i := 0; i < 100; i++ {
		jobs = append(jobs, Job{Id: i})
	}

	var wg sync.WaitGroup
	wg.Add(NumberOfWorkers)
	jobChannel := make(chan Job)
	jobResultChannel := make(chan JobResult, len(jobs))

	// Start the workers
	for i := 0; i < NumberOfWorkers; i++ {
		go worker(i, &wg, jobChannel, jobResultChannel)
	}

	// Send jobs to worker
	for _, job := range jobs {
		jobChannel <- job
	}

	//Now that we've already sent all jobs to the channel, we can close it.
	close(jobChannel)
	wg.Wait()
	//Once wg is completed, since the workers also took care of sending the result to the jobResultChannel, we can close it too.
	close(jobResultChannel)

	var jobResults []JobResult
	// Receive job results from workers
	for result := range jobResultChannel {
		// Read all JobResults from the channel (this is synchronous), and then do whatever you want with the results.
		jobResults = append(jobResults, result)
	}

	// Print all the results
	fmt.Println(jobResults)
	fmt.Printf("Took %s", time.Since(start))
}

/*
our worker function now has a new parameter called jobResultChannel where we'll send our results.
*/
func worker(id int, wg *sync.WaitGroup, jobChannel <-chan Job, resultChannel chan JobResult) {
	defer wg.Done()
	for job := range jobChannel {
		resultChannel <- doSomething(id, job)
	}
}

func doSomething(workerId int, job Job) JobResult {
	fmt.Printf("Worker #%d Running job #%d\n", workerId, job.Id)
	time.Sleep(500 * time.Millisecond)
	return JobResult{Output: "Success"}
}
