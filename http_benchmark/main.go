/*

To avoid usage of mutex I have use following pattern:
* since we know result array size we can construct it dynamically using:
	`make([]T, count)`
* After that I can share pointer on specific element in array to avoid synchronization at all:
	`go do(&arr[i])`

*/

package main

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"sync"
	"time"
)

type URLStatistic struct {
	url           string
	count         int
	total         int64
	timeouts      int
	not2xx        int
	min           int64
	max           int64
	avg           float64
	responseStats []ResponseStatistic
}

func (stat URLStatistic) prettyPrint() {
	fmt.Printf("Test result for URL: %s\n"+
		"Requests count: %d\n"+
		"Total execution time: %d\n"+
		"Timeouts: %d (%.1f%% of total)\n"+
		"Failed requests (not 2xx status code): %d (%.1f%% of total)\n",
		stat.url, stat.count, stat.total,
		stat.timeouts, float64(stat.timeouts)/float64(stat.count)*100,
		stat.not2xx, float64(stat.not2xx)/float64(stat.count)*100)
	if stat.not2xx+stat.timeouts < stat.count {
		fmt.Printf("Min execution time: %d (nanos)\n"+
			"Max execution time: %d (nanos)\n"+
			"Average execution time: %.2f (nanos)\n",
			stat.min, stat.max, stat.avg)
	} else {
		fmt.Println("All requests are failed! Max, min and avg metrics is not available")
	}
	fmt.Println("============================")
}

type ResponseStatistic struct {
	took int64
	err  error
	resp *http.Response
}

func (stat ResponseStatistic) not2xx() bool {
	return stat.resp.StatusCode < 200 || stat.resp.StatusCode >= 300
}

func (stat ResponseStatistic) timeout() bool {
	e, ok := stat.err.(net.Error)
	return ok && e.Timeout()
}

func Benchmark(urls []string, count int, timeout time.Duration) []URLStatistic {
	result := make([]URLStatistic, len(urls))
	client := http.Client{
		Timeout: timeout,
	}
	for i, url := range urls {
		benchURL(client, url, count, &result[i])
	}
	return result
}

func benchURL(client http.Client, url string, count int, urlStats *URLStatistic) {
	responseStats := make([]ResponseStatistic, count)
	var group sync.WaitGroup
	start := time.Now()

	group.Add(count)
	for i := 0; i < count; i++ {
		go makeRequest(client, url, &responseStats[i], &group)
	}
	group.Wait()

	urlStats.total = time.Since(start).Nanoseconds()
	urlStats.url = url
	urlStats.count = count
	urlStats.responseStats = responseStats
	gatherURLStatistic(responseStats, urlStats)
}

func gatherURLStatistic(responseStats []ResponseStatistic, urlStats *URLStatistic) {
	var sum, count int64
	urlStats.max, urlStats.min = math.MinInt64, math.MaxInt64
	for _, responseStat := range responseStats {
		if responseStat.timeout() {
			urlStats.timeouts++
			continue
		}

		if responseStat.not2xx() {
			urlStats.not2xx++
			continue
		}

		sum += responseStat.took
		count++

		if urlStats.max < responseStat.took {
			urlStats.max = responseStat.took
		}
		if urlStats.min > responseStat.took {
			urlStats.min = responseStat.took
		}
	}

	urlStats.avg = float64(sum) / float64(count)
}

func makeRequest(client http.Client, url string, stat *ResponseStatistic, group *sync.WaitGroup) {
	defer group.Done()

	start := time.Now()
	stat.resp, stat.err = client.Get(url)
	stat.took = time.Since(start).Nanoseconds()
}

func main() {
	results := Benchmark([]string{"https://google.com/", "https://vk.com", "http://google.com/404"}, 100, time.Duration(30)*time.Second)
	for _, result := range results {
		result.prettyPrint()
	}
}
