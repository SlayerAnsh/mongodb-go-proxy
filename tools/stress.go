package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	url         = flag.String("url", "http://localhost:8080/api/v1/databases/andromeda-ibc-devnet/collections/ados/documents?limit=5&skip=0", "URL to stress test")
	apiSecret   = flag.String("secret", "readonly-super-secret", "api-secret header value")
	concurrency = flag.Int("c", 10, "Number of concurrent requests")
	duration    = flag.Duration("d", 30*time.Second, "Duration of the test")
	requests    = flag.Int("n", 0, "Total number of requests (0 = run for duration)")
	timeout     = flag.Duration("timeout", 10*time.Second, "Request timeout")
)

type Stats struct {
	totalRequests   int64
	successRequests int64
	failedRequests  int64
	totalDuration   time.Duration
	minDuration     time.Duration
	maxDuration     time.Duration
	statusCodes     map[int]int64
	mu              sync.Mutex
}

func NewStats() *Stats {
	return &Stats{
		statusCodes: make(map[int]int64),
		minDuration: time.Hour, // Initialize with a large value
	}
}

func (s *Stats) RecordRequest(duration time.Duration, statusCode int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	atomic.AddInt64(&s.totalRequests, 1)

	if err != nil || statusCode >= 400 {
		atomic.AddInt64(&s.failedRequests, 1)
	} else {
		atomic.AddInt64(&s.successRequests, 1)
	}

	s.totalDuration += duration
	if duration < s.minDuration {
		s.minDuration = duration
	}
	if duration > s.maxDuration {
		s.maxDuration = duration
	}

	s.statusCodes[statusCode]++
}

func (s *Stats) Print(testDuration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := atomic.LoadInt64(&s.totalRequests)
	success := atomic.LoadInt64(&s.successRequests)
	failed := atomic.LoadInt64(&s.failedRequests)

	if total == 0 {
		fmt.Println("No requests completed")
		return
	}

	avgDuration := s.totalDuration / time.Duration(total)
	successRate := float64(success) / float64(total) * 100

	fmt.Println("\n=== Stress Test Results ===")
	fmt.Printf("Total Requests:     %d\n", total)
	fmt.Printf("Successful:         %d (%.2f%%)\n", success, successRate)
	fmt.Printf("Failed:             %d (%.2f%%)\n", failed, 100-successRate)
	fmt.Printf("\nResponse Times:\n")
	fmt.Printf("  Average:          %v\n", avgDuration)
	fmt.Printf("  Min:              %v\n", s.minDuration)
	fmt.Printf("  Max:              %v\n", s.maxDuration)
	fmt.Printf("\nStatus Codes:\n")
	for code, count := range s.statusCodes {
		fmt.Printf("  %d: %d\n", code, count)
	}

	if testDuration > 0 && total > 0 {
		requestsPerSec := float64(total) / testDuration.Seconds()
		fmt.Printf("\nRequests per second: %.2f\n", requestsPerSec)
	}
}

func makeRequest(client *http.Client, url, apiSecret string, stats *Stats) {
	start := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		stats.RecordRequest(time.Since(start), 0, err)
		return
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-secret", apiSecret)

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		stats.RecordRequest(duration, 0, err)
		return
	}
	defer resp.Body.Close()

	// Read response body to ensure connection is fully processed
	io.Copy(io.Discard, resp.Body)

	stats.RecordRequest(duration, resp.StatusCode, nil)
}

func runStressTest() {
	stats := NewStats()
	client := &http.Client{
		Timeout: *timeout,
	}

	var wg sync.WaitGroup
	stopChan := make(chan struct{})
	startTime := time.Now()

	// Start workers
	var totalRequestCount int64
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					if *requests > 0 {
						current := atomic.LoadInt64(&totalRequestCount)
						if current >= int64(*requests) {
							return
						}
						atomic.AddInt64(&totalRequestCount, 1)
					}
					makeRequest(client, *url, *apiSecret, stats)
				}
			}
		}()
	}

	// Run for specified duration or number of requests
	if *requests > 0 {
		// Wait for all requests to complete
		for atomic.LoadInt64(&totalRequestCount) < int64(*requests) {
			time.Sleep(10 * time.Millisecond)
		}
		close(stopChan)
		wg.Wait()
	} else {
		// Run for specified duration
		time.Sleep(*duration)
		close(stopChan)
		wg.Wait()
	}

	testDuration := time.Since(startTime)
	stats.Print(testDuration)
}

func main() {
	flag.Parse()

	fmt.Printf("Starting stress test...\n")
	fmt.Printf("URL: %s\n", *url)
	fmt.Printf("Concurrency: %d\n", *concurrency)
	if *requests > 0 {
		fmt.Printf("Total Requests: %d\n", *requests)
	} else {
		fmt.Printf("Duration: %v\n", *duration)
	}
	fmt.Printf("Request Timeout: %v\n", *timeout)
	fmt.Println()

	runStressTest()
}
