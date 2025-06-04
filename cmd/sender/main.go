package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	api "github.com/ihatemodels/alcatraz-rest/internal/api/v1"
)

// LoadBalancerStats holds statistics about load balancer distribution
type LoadBalancerStats struct {
	AvailableNodes  int                `json:"available_nodes"`
	TotalRequests   int                `json:"total_requests"`
	SuccessfulReqs  int                `json:"successful_requests"`
	FailedRequests  int                `json:"failed_requests"`
	AverageRespTime int64              `json:"average_response_time_ms"`
	NodeHostnames   []string           `json:"node_hostnames"`
	RequestsPerNode map[string]int     `json:"requests_per_node"`
	ResponseTimes   map[string][]int64 `json:"-"` // Exclude from JSON output
}

// LoadBalancerSummary holds summary statistics for JSON output (without detailed response times)
type LoadBalancerSummary struct {
	AvailableNodes  int            `json:"available_nodes"`
	TotalRequests   int            `json:"total_requests"`
	SuccessfulReqs  int            `json:"successful_requests"`
	FailedRequests  int            `json:"failed_requests"`
	AverageRespTime int64          `json:"average_response_time_ms"`
	NodeHostnames   []string       `json:"node_hostnames"`
	RequestsPerNode map[string]int `json:"requests_per_node"`
}

// SenderConfig holds configuration for the sender application
type SenderConfig struct {
	LoadBalancerURL string
	RequestCount    int
	Concurrency     int
	Timeout         time.Duration
}

var version string

func main() {
	// Initialize simple logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting...", "application", "alcatraz-rest-sender", "version", version)

	// Parse command line flags for sender-specific configuration
	senderCfg := parseSenderFlags()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: senderCfg.Timeout,
	}

	// Send requests and collect statistics
	stats, err := sendRequests(logger, client, senderCfg)
	if err != nil {
		logger.Error("failed to send requests", "error", err)
		os.Exit(1)
	}

	// Display results
	displayResults(logger, stats)
}

func parseSenderFlags() *SenderConfig {
	var (
		url         = flag.String("url", "http://localhost:8080", "Load balancer URL")
		reqCount    = flag.Int("requests", 100, "Number of requests to send")
		concurrency = flag.Int("concurrency", 10, "Number of concurrent requests")
		timeout     = flag.Duration("timeout", 5*time.Second, "Request timeout")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		fmt.Println("Load Balancer Testing Tool")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	return &SenderConfig{
		LoadBalancerURL: *url,
		RequestCount:    *reqCount,
		Concurrency:     *concurrency,
		Timeout:         *timeout,
	}
}

func sendRequests(logger *slog.Logger, client *http.Client, cfg *SenderConfig) (*LoadBalancerStats, error) {
	stats := &LoadBalancerStats{
		RequestsPerNode: make(map[string]int),
		ResponseTimes:   make(map[string][]int64),
	}

	// Channel to limit concurrency
	semaphore := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	logger.Info("starting load balancer test",
		"url", cfg.LoadBalancerURL,
		"requests", cfg.RequestCount,
		"concurrency", cfg.Concurrency)

	startTime := time.Now()

	for i := 0; i < cfg.RequestCount; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			reqStart := time.Now()
			resp, err := client.Get(cfg.LoadBalancerURL + "/api/ping")
			reqDuration := time.Since(reqStart).Milliseconds()

			mu.Lock()
			defer mu.Unlock()

			stats.TotalRequests++

			if err != nil {
				logger.Debug("request failed", "request", reqNum, "error", err)
				stats.FailedRequests++
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Debug("request returned non-200 status",
					"request", reqNum,
					"status", resp.StatusCode)
				stats.FailedRequests++
				return
			}

			var pingResp api.PingResponse
			if err := json.NewDecoder(resp.Body).Decode(&pingResp); err != nil {
				logger.Debug("failed to decode response", "request", reqNum, "error", err)
				stats.FailedRequests++
				return
			}

			// Update statistics
			stats.SuccessfulReqs++
			stats.RequestsPerNode[pingResp.Hostname]++

			if stats.ResponseTimes[pingResp.Hostname] == nil {
				stats.ResponseTimes[pingResp.Hostname] = make([]int64, 0)
			}
			stats.ResponseTimes[pingResp.Hostname] = append(stats.ResponseTimes[pingResp.Hostname], reqDuration)

			logger.Debug("request completed",
				"request", reqNum,
				"hostname", pingResp.Hostname,
				"response_time_ms", reqDuration)

		}(i + 1)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate final statistics
	finalizeStats(stats)

	logger.Info("load balancer test completed",
		"total_duration", totalDuration,
		"total_requests", stats.TotalRequests,
		"successful_requests", stats.SuccessfulReqs,
		"failed_requests", stats.FailedRequests,
		"available_nodes", stats.AvailableNodes)

	return stats, nil
}

func finalizeStats(stats *LoadBalancerStats) {
	// Extract unique hostnames and sort them
	hostnameSet := make(map[string]bool)
	for hostname := range stats.RequestsPerNode {
		hostnameSet[hostname] = true
	}

	stats.NodeHostnames = make([]string, 0, len(hostnameSet))
	for hostname := range hostnameSet {
		stats.NodeHostnames = append(stats.NodeHostnames, hostname)
	}
	sort.Strings(stats.NodeHostnames)

	stats.AvailableNodes = len(stats.NodeHostnames)

	// Calculate average response time
	var totalResponseTime int64
	var totalResponseCount int64

	for _, responseTimes := range stats.ResponseTimes {
		for _, responseTime := range responseTimes {
			totalResponseTime += responseTime
			totalResponseCount++
		}
	}

	if totalResponseCount > 0 {
		stats.AverageRespTime = totalResponseTime / totalResponseCount
	}
}

func displayResults(logger *slog.Logger, stats *LoadBalancerStats) {
	fmt.Println("\n=== Load Balancer Test Results ===")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", stats.SuccessfulReqs)
	fmt.Printf("Failed Requests: %d\n", stats.FailedRequests)
	fmt.Printf("Available Nodes: %d\n", stats.AvailableNodes)
	fmt.Printf("Average Response Time: %d ms\n\n", stats.AverageRespTime)

	fmt.Println("=== Node Hostnames ===")
	for i, hostname := range stats.NodeHostnames {
		fmt.Printf("%d. %s\n", i+1, hostname)
	}

	fmt.Println("\n=== Requests Per Node ===")
	for _, hostname := range stats.NodeHostnames {
		count := stats.RequestsPerNode[hostname]
		percentage := float64(count) / float64(stats.SuccessfulReqs) * 100
		fmt.Printf("%-20s: %4d requests (%.1f%%)\n", hostname, count, percentage)
	}

	fmt.Println("\n=== Response Time Statistics (ms) ===")
	for _, hostname := range stats.NodeHostnames {
		responseTimes := stats.ResponseTimes[hostname]
		if len(responseTimes) == 0 {
			continue
		}

		var sum, min, max int64
		min = responseTimes[0]
		max = responseTimes[0]

		for _, rt := range responseTimes {
			sum += rt
			if rt < min {
				min = rt
			}
			if rt > max {
				max = rt
			}
		}

		avg := sum / int64(len(responseTimes))
		fmt.Printf("%-20s: avg=%3dms, min=%3dms, max=%3dms, count=%d\n",
			hostname, avg, min, max, len(responseTimes))
	}

	// Output JSON for programmatic use (without detailed response times)
	fmt.Println("\n=== JSON Output ===")
	summary := &LoadBalancerSummary{
		AvailableNodes:  stats.AvailableNodes,
		TotalRequests:   stats.TotalRequests,
		SuccessfulReqs:  stats.SuccessfulReqs,
		FailedRequests:  stats.FailedRequests,
		AverageRespTime: stats.AverageRespTime,
		NodeHostnames:   stats.NodeHostnames,
		RequestsPerNode: stats.RequestsPerNode,
	}
	jsonOutput, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		logger.Error("failed to marshal stats to JSON", "error", err)
	} else {
		fmt.Println(string(jsonOutput))
	}
}
