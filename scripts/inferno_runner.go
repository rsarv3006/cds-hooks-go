package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

const (
	infernoURL = "http://localhost:4567"
	testSuite  = "crd_server"
)

type TestSuite struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	TestCount int     `json:"test_count"`
	Inputs    []Input `json:"inputs"`
}

type Input struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Default string `json:"default,omitempty"`
}

type TestSession struct {
	ID          string `json:"id"`
	TestSuiteID string `json:"test_suite_id"`
	CreatedAt   string `json:"created_at"`
}

type TestRun struct {
	ID            string `json:"id"`
	TestSessionID string `json:"test_session_id,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type Result struct {
	ID          string `json:"id"`
	TestID      string `json:"test_id"`
	TestGroupID string `json:"test_group_id"`
	Result      string `json:"result"`
	Message     string `json:"result_message,omitempty"`
}

func main() {
	baseURL := os.Getenv("CDS_HOOKS_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	fmt.Printf("Testing CDS Hooks server at: %s\n", baseURL)
	fmt.Printf("Inferno at: %s\n", infernoURL)
	fmt.Println()

	presetFile := "/opt/inferno/data/local_preset.json"
	if _, err := os.Stat("/home/rjs/inferno-crd/data/local_preset.json"); err != nil {
		fmt.Printf("Preset file not found: /home/rjs/inferno-crd/data/local_preset.json\n")
		os.Exit(1)
	}

	runGroup := func(group string) {
		fmt.Printf("\n=== Running Group %s ===\n", group)
		cmd := exec.Command("podman", "exec", "inferno-app", "bundle", "exec", "inferno", "execute",
			"--suite", testSuite,
			"--preset-file", presetFile,
			"--groups", group,
			"--outputter", "json")
		cmd.Dir = "/home/rjs/inferno-crd"

		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error running tests: %v\n", err)
			fmt.Printf("Output: %s\n", out.String())
			return
		}

		var results []Result
		json.Unmarshal(out.Bytes(), &results)

		var passed, failed, skipped int
		for _, r := range results {
			if r.TestID == "" {
				continue
			}
			icon := "✓"
			if r.Result == "fail" {
				icon = "✗"
				failed++
			} else if r.Result == "skip" {
				icon = "-"
				skipped++
			} else if r.Result == "pass" {
				passed++
			}
			fmt.Printf("  %s %s: %s\n", icon, r.TestID, r.Result)
			if r.Message != "" && r.Result == "fail" {
				fmt.Printf("      %s\n", r.Message)
			}
		}
		fmt.Printf("\nSummary: Passed: %d, Failed: %d, Skipped: %d\n", passed, failed, skipped)
	}

	runGroup("1")
	runGroup("2")
	runGroup("3")

	fmt.Println("\n=== Done ===")
}
