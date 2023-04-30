### Worker based Rabbitmq Event consumers and dispatchers.

### Running

```sh
# start multiple listners in separate shells
go run listner.go
go run listner.go

# dispatch some events in another shell
go run dispatcher.go

```

### Implementation
- Add custom queue implementations in worker package. 
- Two sample queues are provided [ worker.AsyncJobABCQueue, worker.SendWebhookQueue]

 - Use worker.Run(conn, queues[]) method to start a worker listening to a list of given queues. This is a blocking method which will keep listening until a kill signal.
- Use worker.SomeJobName() func to send a job to rabbitmq queue.
- WORKER_COUNT & RABBITMQ_URL -- env variables are hardcoded in rabbitmq/rabbitmq.go

### Folder structure

```bash
# Package helper
├───connection  # ->  All external connections, including database, httpclient, grpcclient, rabbitmq client etc.
├───rabbitmq    # ->  rabbitmq operations -- manage all interactions with a given rabbitmq connection. WORKER_COUNT & RABBITMQ_URL -- variables
├───signal      # ->  catch os interrupts and exit signal -- handle waitforinterrupt once workers are started asynchronously. (worker.go Run method)
├───worker      # ->  Different worker queue implementations with publish and listner methods for each in its own file. 
                #     worker.go -> Run method starts all listners aysnchronously and waits for os interrupt. Publish method allows publishing to given queue name
└
# listenr.go    --> starts worker.Run method by passing the required Queue names as string to listen to those queues in a goroutine and wait for any interrupts.
# dispatcher.go --> dispatches multiple events into predefined queues

```
