
# bench_chan_buffer_test.go README

> If the capacity is zero or absent, the channel is unbuffered and communication succeeds only when both a sender and receiver are ready.
The documentation of effective Go is also very clear about that:
If the channel is unbuffered, the sender blocks until the receiver has received the value.

>>  The size of the buffer we define during the channel creation could drastically impact the performances. We will use the fan-out pattern that intensively uses the channel in order to see the impact of different buffer sizes.

## Run using
    go test bench_chan_buffer_test.go -bench=.

or single testcase using

    go test bench_chan_buffer_test.go -bench=^BenchmarkWithNoBuffer$

or execute all _test files with 

    go test -bench=.

or print with memory 
    
    go test -bench=. -benchmem

To avoid executing any test functions in the test files, pass a regular expression to the -run flag:
The -run flag is used to specify which unit tests should be executed.
By using ^# as the argument to -run, we effectively filter out all of the unit test functions.

    go test -bench=. -count 5 -run=^#

>>sample o/p <<

```
goos: linux
goarch: amd64
pkg: abidhmuhsin.com/go-snippets/go-channels/bench
cpu: Intel(R) Core(TM) i7-1065G7 CPU @ 1.30GHz
BenchmarkWithNoBuffer-8                                     2038            601221 ns/op
BenchmarkWithBufferSizeOf1-8                                2662            485742 ns/op
BenchmarkWithBufferSizeEqualsToNumberOfWorker-8             3626            308257 ns/op
BenchmarkWithBufferSizeExceedsNumberOfWorker-8              5800            203491 ns/op
PASS
ok      abidhmuhsin.com/go-snippets/go-channels/bench   6.694s
```
goos, goarch, pkg, and cpu describe the operating system, architecture, package, and CPU specifications, respectively.
BenchmarkWithNoBuffer-8 denotes the name of the benchmark function that was run.
The -8 suffix denotes the number of CPUs used to run the benchmark, as specified by GOMAXPROCS.
On the right side of the function name, you have two values, 14588 and 82798 ns/op.
The former indicates the total number of times the loop was executed, while the latter is the average amount of time each iteration took to complete, expressed in nanoseconds per operation.

----------------

In our benchmark, one producer will inject a one million integer element in the channel
while ten workers will read and add them to a single result variable named total.

 > ## A well sized buffer could really make your application faster! 

 <!-- may analyze the traces of our benchmarks to confirm where the latencies are. -->



benchstat usage for analyzing two diff implementations

>  go get golang.org/x/perf/cmd/benchstat

    [for i in {1..20}; do go test -bench=.$ >> old.txt; done]
    [for i in {1..20}; do go test -bench=.$ >> new.txt; done]
or use

    [go test -bench=. -count=2 >> old.txt]
    [go test -bench=. -count=2 >> new.txt]  

    [benchstat new.txt old.txt ]
 >> OUTPUT

```
 name                                    old time/op  new time/op  delta
WithNoBuffer-8                           576µs ± 4%   590µs ± 7%  +2.40%  (p=0.009 n=19+18)
WithBufferSizeOf1-8                      452µs ± 2%   471µs ± 8%  +4.18%  (p=0.000 n=19+18)
WithBufferSizeEqualsToNumberOfWorker-8   296µs ± 2%   305µs ± 9%  +3.27%  (p=0.000 n=18+18)
WithBufferSizeExceedsNumberOfWorker-8    202µs ± 2%   206µs ± 4%  +1.76%  (p=0.000 n=18+20)
```
>> WithBufferSizeExceedsNumberOfWorker -- best performance


## References
- https://medium.com/a-journey-with-go/go-buffered-and-unbuffered-channels-29a107c00268
- https://blog.logrocket.com/benchmarking-golang-improve-function-performance/