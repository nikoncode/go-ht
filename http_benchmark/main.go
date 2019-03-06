package main

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type Statistic struct {
	min      int64
	max      int64
	avg      float64
	timeouts int64
}

func Benchmark(url string, requests int, timeout time.Duration) Statistic {
	var mutex sync.Mutex
	var group sync.WaitGroup
	client := http.Client{
		Timeout: timeout,
	}
	var sum, timeouts, max, min, count int64
	max, min = math.MinInt64, math.MaxInt64

	for i := 0; i < requests; i++ {
		group.Add(1)
		go func() {
			defer group.Done()

			start := time.Now()
			_, err := client.Get(url)
			took := time.Since(start).Nanoseconds()

			//gather statistics
			mutex.Lock()
			defer mutex.Unlock()

			if e, ok := err.(net.Error); ok && e.Timeout() {
				timeouts++
				//do not gather statistics when timeout reached
				return
			} else if err != nil {
				panic(err)
			}

			sum += took
			count++
			if max < took {
				max = took
			}
			if min > took {
				min = took
			}
		}()
	}
	group.Wait()
	return Statistic{min: min, max: max, avg: float64(sum) / float64(count), timeouts: timeouts}
}

func main() {
	fmt.Printf("%+v", Benchmark("https://google.com/", 100, time.Duration(30)*time.Second))
}
