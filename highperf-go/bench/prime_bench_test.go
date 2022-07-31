// https://blog.logrocket.com/benchmarking-golang-improve-function-performance/

package bench

import (
	"fmt"
	"math"
	"testing"
)

var table = []struct {
	input int
}{
	{input: 100},
	{input: 1000},
	{input: 74382},
	{input: 382399},
}

func BenchmarkPrimeNumbers(b *testing.B) {
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size_%d", v.input), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// primeNumbers(v.input)        //old
				sieveOfEratosthenes(v.input) //new

			}
		})
	}
}

/*
func BenchmarkSieveOfErastosthenes(b *testing.B) {
	for _, v := range table {
		b.Run(fmt.Sprintf("input_size_%d", v.input), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sieveOfEratosthenes(v.input)
			}
		})
	}
}
*/

func primeNumbers(max int) []int {
	var primes []int

	for i := 2; i < max; i++ {
		isPrime := true

		for j := 2; j <= int(math.Sqrt(float64(i))); j++ {
			if i%j == 0 {
				isPrime = false
				break
			}
		}

		if isPrime {
			primes = append(primes, i)
		}
	}

	return primes
}

func sieveOfEratosthenes(max int) []int {
	b := make([]bool, max)

	var primes []int

	for i := 2; i < max; i++ {
		if b[i] {
			continue
		}

		primes = append(primes, i)

		for k := i * i; k < max; k += i {
			b[k] = true
		}
	}

	return primes
}
