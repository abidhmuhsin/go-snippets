package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	Id int
}

type JobResult struct {
	Output string
}

const (
	NumberOfWorkers                    = 10
	MaximumNumberOfExecutionsPerSecond = 50
)

func main() {
	start := time.Now()
	// Create fake jobs for testing purposes
	var jobs []Job
	for i := 0; i < 500; i++ {
		jobs = append(jobs, Job{Id: i})
	}

	var (
		wg         sync.WaitGroup
		jobChannel = make(chan Job)
	)
	wg.Add(NumberOfWorkers)

	// Start the workers
	for i := 0; i < NumberOfWorkers; i++ {
		go worker(i, &wg, jobChannel)
	}

	// Send jobs to worker
	for _, job := range jobs {
		jobChannel <- job
	}

	close(jobChannel)
	wg.Wait()
	fmt.Printf("Took %s\n", time.Since(start))
}

/*
What this does is divide the total number of workers by the maximum amount of calls per second,
and throttles each worker to what the result of that equation is.
It's not the most accurate way of doing it, since some workers may get several jobs with long
 execution time in a row while other workers may get several jobs with short
 execution time in a row, but it's the easiest way of doing it and it's generally good enough.

In other words, if you have NumberOfWorkers set to 2 and MaximumNumberOfExecutionsPerSecond set to 10,
 each workers would be throttled to calling doSomething() at most once every 200ms.
*/
func worker(id int, wg *sync.WaitGroup, jobChannel <-chan Job) {
	defer wg.Done()
	lastExecutionTime := time.Now()
	minimumTimeBetweenEachExecution := time.Duration(math.Ceil(1e9 / (MaximumNumberOfExecutionsPerSecond / float64(NumberOfWorkers))))
	for job := range jobChannel {
		timeUntilNextExecution := -(time.Since(lastExecutionTime) - minimumTimeBetweenEachExecution)
		if timeUntilNextExecution > 0 {
			fmt.Printf("Worker #%d backing off for %s\n", id, timeUntilNextExecution.String())
			time.Sleep(timeUntilNextExecution)
		} else {
			fmt.Printf("Worker #%d not backing off\n", id)
		}
		lastExecutionTime = time.Now()
		doSomething(id, job)
	}
}

func doSomething(workerId int, job Job) JobResult {
	simulatedExecutionTime := rand.Intn(1000)
	fmt.Printf("Worker #%d Running job #%d (simulatedExecutionTime=%dms)\n", workerId, job.Id, simulatedExecutionTime)
	time.Sleep(time.Duration(simulatedExecutionTime) * time.Millisecond)
	return JobResult{Output: "Success"}
}
