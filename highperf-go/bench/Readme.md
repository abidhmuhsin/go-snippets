## Run using
    go test prime_bench_test.go -bench=.

or single testcase using

    go test prime_bench_test.go -bench=^BenchmarkPrimeNumbers$

or execute all _test files with 

    go test -bench=.

or print with memory 
    
    go test -bench=. -benchmem

To avoid executing any test functions in the test files, pass a regular expression to the -run flag:
The -run flag is used to specify which unit tests should be executed.
By using ^# as the argument to -run, we effectively filter out all of the unit test functions.

    go test -bench=. -count 5 -run=^#

> run

    go test prime_bench_test.go -bench=Prime
> o/p

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-1065G7 CPU @ 1.30GHz
BenchmarkPrimeNumbers/input_size_100-8            473031              2371 ns/op
BenchmarkPrimeNumbers/input_size_1000-8            26289             45091 ns/op
BenchmarkPrimeNumbers/input_size_74382-8              76          15593294 ns/op
BenchmarkPrimeNumbers/input_size_382399-8              7         146904230 ns/op
PASS
ok      command-line-arguments  5.184s
```
goos, goarch, pkg, and cpu describe the operating system, architecture, package, and CPU specifications, respectively.
BenchmarkWithNoBuffer-8 denotes the name of the benchmark function that was run.
The -8 suffix denotes the number of CPUs used to run the benchmark, as specified by GOMAXPROCS.
On the right side of the function name, you have two values, 14588 and 82798 ns/op.
The former indicates the total number of times the loop was executed, while the latter is the average amount of time each iteration took to complete, expressed in nanoseconds per operation.


> #### Upon first glance, we can see that the Sieve of Eratosthenes algorithm is much more performant than the previous algorithm. However, instead of eyeballing the results to compare the performance between runs, we can use a tool like benchstat, which helps us compute and compare benchmarking statistics.

>> Now comment sieveOfEratosthenes(v.input)  and uncomment  primeNumbers(v.input)and then run

>     go test prime_bench_test.go -bench=Prime -count 1 | tee normal_prime.txt
```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-1065G7 CPU @ 1.30GHz
BenchmarkPrimeNumbers/input_size_100-8            507060              2285 ns/op
BenchmarkPrimeNumbers/input_size_1000-8            26144             46633 ns/op
BenchmarkPrimeNumbers/input_size_74382-8              76          15631181 ns/op
BenchmarkPrimeNumbers/input_size_382399-8              7         147286186 ns/op
PASS
ok      command-line-arguments  6.188s
```

>> Now comment  primeNumbers(v.input) and uncomment sieveOfEratosthenes(v.input) and then run

>     go test prime_bench_test.go -bench=Prime -count 1 | tee sieveOfEratosthenes.txt
```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i7-1065G7 CPU @ 1.30GHz
BenchmarkSieveOfErastosthenes/input_size_100-8            943540              1060 ns/op
BenchmarkSieveOfErastosthenes/input_size_1000-8           199666              7145 ns/op
BenchmarkSieveOfErastosthenes/input_size_74382-8            2164            544122 ns/op
BenchmarkSieveOfErastosthenes/input_size_382399-8            358           2978864 ns/op
PASS
ok      command-line-arguments  7.001s
```

> benchstat normal_prime.txt sieveOfEratosthenes.txt


```
name                              normal_prime time/op  sieveOfEratosthenes time/op  delta
PrimeNumbers/input_size_100-8     2.40µs ± 0%  0.97µs ± 0%   ~     (p=1.000 n=1+1)
PrimeNumbers/input_size_1000-8    45.4µs ± 0%   6.0µs ± 0%   ~     (p=1.000 n=1+1)
PrimeNumbers/input_size_74382-8   15.6ms ± 0%   0.6ms ± 0%   ~     (p=1.000 n=1+1)
PrimeNumbers/input_size_382399-8   148ms ± 0%     3ms ± 0%   ~     (p=1.000 n=1+1)
```

The delta column reports the percentage change in performance, the P-value, and the number of samples that are considered to be valid, n. If you see an n value lower than the number of samples taken, it may mean that your environment wasn’t stable enough while the samples were being collected. See the benchstat docs to see the other options available to you.