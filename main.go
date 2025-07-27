package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type EmailConfig struct {
	SMTPHost  string `json:"smtp_host"`
	SMTPPort  int    `json:"smtp_port"`
	Sender    string `json:"sender"`
	Password  string `json:"password"`
	Recipient string `json:"recipient"`
}

type Config struct {
	Websites []string    `json:"websites"`
	Email    EmailConfig `json:"email"`
}

func loadConfiguration(file string) (Config, error) {
	var config Config
	configFile, err := os.Open(file)
	if err != nil {
		return config, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

var statusMap = make(map[string]string)
var statusMutex = &sync.Mutex{}

func sendEmail(emailConfig EmailConfig, url string) {
	auth := smtp.PlainAuth("", emailConfig.Sender, emailConfig.Password, emailConfig.SMTPHost)
	to := []string{emailConfig.Recipient}
	msg := []byte("To: " + emailConfig.Recipient + "\r\n" +
		"Subject: Website Down: " + url + "\r\n" +
		"\r\n" +
		"The website " + url + " is currently down.\r\n")

	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	err := smtp.SendMail(addr, auth, emailConfig.Sender, to, msg)
	if err != nil {
		fmt.Printf("Error sending email for %s: %s\n", url, err)
		fmt.Println("Please ensure your email settings in config.json are correct.")
		return
	}
	fmt.Printf("Email notification sent for %s\n", url)
}

func checkWebsite(url string, emailConfig EmailConfig) {
	resp, err := http.Get(url)
	statusMutex.Lock()
	defer statusMutex.Unlock()
	lastStatus := statusMap[url]

	if err != nil {
		fmt.Printf("Website %s is down: %s\n", url, err)
		if lastStatus != "down" {
			sendEmail(emailConfig, url)
			statusMap[url] = "down"
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		fmt.Printf("Website %s is up. Status: %s\n", url, resp.Status)
		statusMap[url] = "up"
	} else {
		fmt.Printf("Website %s is down. Status: %s\n", url, resp.Status)
		if lastStatus != "down" {
			sendEmail(emailConfig, url)
			statusMap[url] = "down"
		}
	}
}

func startMonitoring(config Config) {
	// Initial check
	fmt.Println("--- Initial Check ---")
	var wg sync.WaitGroup
	for _, site := range config.Websites {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			checkWebsite(url, config.Email)
		}(site)
	}
	wg.Wait()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("\n--- New Check Cycle ---")
			var wg sync.WaitGroup
			for _, site := range config.Websites {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					checkWebsite(url, config.Email)
				}(site)
			}
			wg.Wait()
		}
	}
}

type StatusEntry struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type PaginatedStatusResponse struct {
	TotalPages  int           `json:"totalPages"`
	CurrentPage int           `json:"currentPage"`
	Data        []StatusEntry `json:"data"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	statusMutex.Lock()
	// Convert map to slice for sorting and pagination
	var statuses []StatusEntry
	for url, status := range statusMap {
		statuses = append(statuses, StatusEntry{URL: url, Status: status})
	}
	statusMutex.Unlock()

	// Sort by URL for consistent ordering
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].URL < statuses[j].URL
	})

	totalItems := len(statuses)
	totalPages := (totalItems + limit - 1) / limit

	start := (page - 1) * limit
	end := start + limit
	if start > totalItems {
		start = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	paginatedData := statuses[start:end]

	response := PaginatedStatusResponse{
		TotalPages:  totalPages,
		CurrentPage: page,
		Data:        paginatedData,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development, allow any origin
	json.NewEncoder(w).Encode(response)
}

func startAPIServer() {
	http.HandleFunc("/status", statusHandler)
	fmt.Println("API server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting API server:", err)
	}
}

func main() {
	fmt.Println("Uptime Monitor Starting...")
	config, err := loadConfiguration("config.json")
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	go startAPIServer()
	startMonitoring(config)
}
