package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type job func() // тип функция

func worker(id int, jobs <-chan job, wg *sync.WaitGroup) {
	defer wg.Done()
	for fn := range jobs {
		//fmt.Println("worker", id, "started  job")
		fn()
		//fmt.Println("worker", id, "finished job")
		//results <- struct{}{} // sending empty struct to signal "finished"
	}
}

func main() {
	current := time.Now()
	dosome()
	fmt.Println("Total time: ", time.Since(current))

}

//func ratelimiter(ch chan job, totalWorker int, ratePerM int) {

func dosome() {
	var wg sync.WaitGroup
	const numJobs = 6
	jobs := make(chan job)
	//results := make(chan struct{}, numJobs)

	rand.Seed(time.Now().UnixNano() + rand.Int63())
	sliceJobs := make([]job, 3) // Слайс функций для нагрузки
	sliceJobs[0] = func() {     // одну переменный типа job
		current := time.Now()
		time.Sleep(time.Duration(rand.Intn(1e2)) * time.Millisecond)
		fmt.Println("Hello", time.Since(current))
	}

	sliceJobs[1] = func() { // вторую
		current := time.Now()
		time.Sleep(time.Duration(rand.Intn(2e2)) * time.Millisecond)
		fmt.Println("Golang", time.Since(current))
	}
	sliceJobs[2] = func() { // и еще одну
		current := time.Now()
		time.Sleep(time.Duration(rand.Intn(3e2)) * time.Millisecond)
		fmt.Println("World", time.Since(current))
	}

	for w := 1; w <= 5; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg)
	}

	for i := 0; i < 10; i++ {
		jobs <- sliceJobs[rand.Intn(3)]
	}

	close(jobs)
	wg.Wait()

}
