// Problem 2 - Even Fibonacci numbers - https://projecteuler.net/problem=2
//Each new term in the Fibonacci sequence is generated by adding the previous two terms. By starting with 1 and 2, the first 10 terms will be:
//1, 2, 3, 5, 8, 13, 21, 34, 55, 89, ...
//By considering the terms in the Fibonacci sequence whose values do not exceed four million, find the sum of the even-valued terms.
package main

import "fmt"

func main() {
	sum, next := 2, 0
	fib := []int{1, 2}
	for {
		next = fib[0] + fib[1]
		if next > 4000000 {
			break
		}

		if next%2 == 0 {
			sum += next
		}
		fib = append(fib[1:], next)
	}
	fmt.Println(sum)
}