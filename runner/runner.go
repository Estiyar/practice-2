package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8080"
const apiKey = "secret12345"

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UpdateResponse struct {
	Updated bool `json:"updated"`
}

func main() {
	failed := false

	if !testServerAvailability() {
		failed = true
	}

	if !testCreateTask() {
		failed = true
	}

	if !testGetTaskByID() {
		failed = true
	}

	if !testUpdateTask() {
		failed = true
	}

	if !testValidationErrors() {
		failed = true
	}

	if !testAuthentication() {
		failed = true
	}

	if !testRateLimiting() {
		failed = true
	}

	if failed {
		os.Exit(1)
	}

	println("All tests passed")
	os.Exit(0)
}

func testServerAvailability() bool {
	println("Testing server availability...")

	resp, err := makeRequest("GET", "/tasks", nil, true)
	if err != nil {
		println("FAIL: Server not available:", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		println("FAIL: Expected 200, got", resp.StatusCode)
		return false
	}

	println("PASS: Server availability")
	return true
}

func testCreateTask() bool {
	println("Testing task creation...")

	body := map[string]string{"title": "Test Task"}
	resp, err := makeRequest("POST", "/tasks", body, true)
	if err != nil {
		println("FAIL: Create task request failed:", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		println("FAIL: Expected 201, got", resp.StatusCode)
		return false
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		println("FAIL: Failed to decode response:", err.Error())
		return false
	}

	if task.ID == 0 || task.Title != "Test Task" || task.Done != false {
		println("FAIL: Invalid task data")
		return false
	}

	println("PASS: Task creation")
	return true
}

func testGetTaskByID() bool {
	println("Testing get task by ID...")

	body := map[string]string{"title": "Task For Get"}
	resp, err := makeRequest("POST", "/tasks", body, true)
	if err != nil {
		println("FAIL: Create task failed:", err.Error())
		return false
	}

	var task Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	resp, err = makeRequest("GET", fmt.Sprintf("/tasks?id=%d", task.ID), nil, true)
	if err != nil {
		println("FAIL: Get task request failed:", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		println("FAIL: Expected 200, got", resp.StatusCode)
		return false
	}

	var retrieved Task
	if err := json.NewDecoder(resp.Body).Decode(&retrieved); err != nil {
		println("FAIL: Failed to decode response:", err.Error())
		return false
	}

	if retrieved.ID != task.ID {
		println("FAIL: Task ID mismatch")
		return false
	}

	println("PASS: Get task by ID")
	return true
}

func testUpdateTask() bool {
	println("Testing task update...")

	createBody := map[string]string{"title": "Task For Update"}
	resp, err := makeRequest("POST", "/tasks", createBody, true)
	if err != nil {
		println("FAIL: Create task failed:", err.Error())
		return false
	}

	var task Task
	json.NewDecoder(resp.Body).Decode(&task)
	resp.Body.Close()

	updateBody := map[string]bool{"done": true}
	resp, err = makeRequest("PATCH", fmt.Sprintf("/tasks?id=%d", task.ID), updateBody, true)
	if err != nil {
		println("FAIL: Update task request failed:", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		println("FAIL: Expected 200, got", resp.StatusCode)
		return false
	}

	var updateResp UpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		println("FAIL: Failed to decode response:", err.Error())
		return false
	}

	if !updateResp.Updated {
		println("FAIL: Update response invalid")
		return false
	}

	println("PASS: Task update")
	return true
}

func testValidationErrors() bool {
	println("Testing validation errors...")

	emptyTitleBody := map[string]string{"title": ""}
	resp, err := makeRequest("POST", "/tasks", emptyTitleBody, true)
	if err != nil {
		println("FAIL: Empty title request failed:", err.Error())
		return false
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		println("FAIL: Expected 400 for empty title, got", resp.StatusCode)
		return false
	}

	updateBody := map[string]bool{"done": true}
	resp, err = makeRequest("PATCH", "/tasks", updateBody, true)
	if err != nil {
		println("FAIL: Missing ID request failed:", err.Error())
		return false
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		println("FAIL: Expected 400 for missing ID, got", resp.StatusCode)
		return false
	}

	println("PASS: Validation errors")
	return true
}

func testAuthentication() bool {
	println("Testing authentication...")

	resp, err := makeRequest("GET", "/tasks", nil, false)
	if err != nil {
		println("FAIL: No auth request failed:", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		println("FAIL: Expected 401, got", resp.StatusCode)
		return false
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		println("FAIL: Failed to decode error response:", err.Error())
		return false
	}

	if errResp.Error != "unauthorized" {
		println("FAIL: Expected 'unauthorized' error message")
		return false
	}

	println("PASS: Authentication")
	return true
}

func testRateLimiting() bool {
	println("Testing rate limiting...")

	hitLimit := false
	body := map[string]string{"title": "Rate Limit Test"}

	for i := 0; i < 15; i++ {
		resp, err := makeRequest("POST", "/tasks", body, true)
		if err != nil {
			println("FAIL: Rate limit request failed:", err.Error())
			return false
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			var errResp ErrorResponse
			json.NewDecoder(resp.Body).Decode(&errResp)
			resp.Body.Close()

			if errResp.Error == "rate limit exceeded" {
				hitLimit = true
				break
			}
		}
		resp.Body.Close()
	}

	if !hitLimit {
		println("FAIL: Rate limit not triggered")
		return false
	}

	time.Sleep(1100 * time.Millisecond)

	resp, err := makeRequest("POST", "/tasks", body, true)
	if err != nil {
		println("FAIL: Post-limit request failed:", err.Error())
		return false
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		println("FAIL: Rate limit did not reset after 1 second")
		return false
	}

	println("PASS: Rate limiting")
	return true
}

func makeRequest(method, path string, body interface{}, withAuth bool) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if withAuth {
		req.Header.Set("X-API-KEY", apiKey)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	compactBody := bytes.Buffer{}
	if json.Compact(&compactBody, bodyBytes) == nil {
		fmt.Printf("%s %s -> %d %s\n", method, path, resp.StatusCode, compactBody.String())
	} else {
		fmt.Printf("%s %s -> %d %s\n", method, path, resp.StatusCode, string(bodyBytes))
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, nil
}