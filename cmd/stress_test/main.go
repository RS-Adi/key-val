package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	url := "http://localhost:8080"
	var wg sync.WaitGroup
	requests := 100

	fmt.Printf("Starting stress test with %d concurrent requests...\n", requests)
	start := time.Now()

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			val := fmt.Sprintf("val-%d", id)

			// SET
			resp, err := http.Post(
				fmt.Sprintf("%s/set", url),
				"application/json",
				strings.NewReader(fmt.Sprintf(`{"key":"%s","value":"%s"}`, key, val)),
			)
			if err != nil {
				fmt.Printf("Request %d SET failed: %v\n", id, err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Request %d SET status: %d\n", id, resp.StatusCode)
				return
			}

			// GET
			resp, err = http.Get(fmt.Sprintf("%s/get?key=%s", url, key))
			if err != nil {
				fmt.Printf("Request %d GET failed: %v\n", id, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Request %d GET status: %d\n", id, resp.StatusCode)
				return
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Stress test completed in %v\n", duration)
}
