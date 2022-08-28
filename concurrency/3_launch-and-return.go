package main

import (
	"fmt"
	"sync"
)

type Job struct {
	Id int
}

type JobResult struct {
	Output string
}

func main() {
	var jobs []Job
	for i := 0; i < 5; i++ {
		jobs = append(jobs, Job{Id: i})
	}
	jobResults := launchAndReturn(jobs)
	fmt.Println(jobResults)
}

func launchAndReturn(jobs []Job) []JobResult {
	var (
		results      []JobResult
		resultsMutex = sync.RWMutex{} //We'd be risking making concurrent changes to a slice, which would cause problems. To prevent that, we'll use sync.RWMutex, which will allow us to prevent concurrent changes.
		wg           sync.WaitGroup   //In order to return the result of multiple jobs, we need to make sure to wait that all jobs are done before returning.
	)
	wg.Add(len(jobs)) //sync.WaitGroup, will allow us to add the number of jobs that needs to be completed to the WaitGroup and keep track of when all jobs are completed.

	//iterate over each job, and start a new goroutine for each entry.
	for _, job := range jobs {
		go func(job Job) {
			defer wg.Done() //To make sure that the job is marked as done at the end of each goroutines, we'll defer wg.Done() right now.
			jobResult := doSomething(job)
			resultsMutex.Lock()
			results = append(results, jobResult)
			resultsMutex.Unlock()
		}(job)
	}

	//The last step is to wait for the WaitGroup to complete by using wg.Wait(), and then we can return the results.
	wg.Wait()
	return results
}

func doSomething(job Job) JobResult {
	fmt.Printf("Running job #%d\n", job.Id)
	return JobResult{Output: "Success"}
}
