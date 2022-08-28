package main

// throttle:- maximum number of concurrent workers.
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

	var (
		wg         sync.WaitGroup
		jobChannel = make(chan Job)
	)
	wg.Add(NumberOfWorkers)

	// Start the workers
	for i := 0; i < NumberOfWorkers; i++ {
		// Now that our channel is ready, we can start the workers. For now, they'll just sit here and wait, since we haven't sent the jobs to the channel yet.
		go worker(i, &wg, jobChannel)
	}

	// Send jobs to worker
	for _, job := range jobs {
		//We send the jobs to the channel here, which is being read on by our workers.
		jobChannel <- job
	}

	//Since we already sent each jobs to the channel, we can close the channel and wait for the workers to finish.
	close(jobChannel)
	wg.Wait()
	fmt.Printf("Took %s\n", time.Since(start))
}

/* The worker (which is just a function) will be launched on a different goroutine,
 and will listen to the channel for new jobs available:
Note that a channel is like a pointer to a queue, meaning that two workers who are reading from the same queue won't get the same jobs.
One job will be received by only one worker.
*/
func worker(id int, wg *sync.WaitGroup, jobChannel <-chan Job) {
	defer wg.Done()
	// Iterating over a channel means you're listening to the channel, and this listening will continue until the channel is closed. Note that multiple workers will be listening to the same channel, Go channels is made for the producer-consumer problem.
	for job := range jobChannel {
		doSomething(id, job)
	}
}

func doSomething(workerId int, job Job) JobResult {
	fmt.Printf("Worker #%d Running job #%d\n", workerId, job.Id)
	time.Sleep(1 * time.Second)
	return JobResult{Output: "Success"}
}
