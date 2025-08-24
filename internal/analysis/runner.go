package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunPa11y executes the pa11y command and returns the result.
func RunPa11y(url string, runner string) ([]Issue, error) {
	if runner == "" {
		runner = "htmlcs"
	}

	// Determine the pa11y command to run, allowing override via PA11Y_COMMAND (e.g., "npx pa11y").
	pa11yCmd := os.Getenv("PA11Y_COMMAND")
	if strings.TrimSpace(pa11yCmd) == "" {
		pa11yCmd = "pa11y"
	}
	parts := strings.Fields(pa11yCmd)
	var execName string
	var baseArgs []string
	if len(parts) > 0 {
		execName = parts[0]
		if len(parts) > 1 {
			baseArgs = parts[1:]
		}
	} else {
		execName = "pa11y"
	}

	args := append([]string{}, baseArgs...)
	args = append(args, "--reporter", "json", "--runner", runner, url)

	cmd := exec.Command(execName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// pa11y exits with code 2 if there are accessibility issues.
		// We still want to see the JSON report in that case.
		// We only exit if there's a different error.
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 2 {
				return nil, fmt.Errorf("error running pa11y: %v\nOutput: %s", err, output)
			}
		} else {
			return nil, fmt.Errorf("error running pa11y: %v\nOutput: %s", err, output)
		}
	}

	var result []Issue
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pa11y output: %v\nOutput was: %s", err, output)
	}

	return result, nil
}

// Worker processes analysis tasks from the queue.
type Worker struct {
	service *Service
}

// NewWorker creates a new worker.
func NewWorker(service *Service) *Worker {
	return &Worker{
		service: service,
	}
}

// Start begins the worker's processing loop.
func (w *Worker) Start() {
	go func() {
		for {
			analysisID := w.service.GetNextFromQueue()
			analysis, ok := w.service.GetByID(analysisID)
			if !ok {
				fmt.Fprintf(os.Stderr, "Error: analysis with ID %s not found\n", analysisID)
				continue
			}

			w.service.UpdateStatus(analysis.ID, StatusProcessing)

			// for now, we'll use the default runner
			result, err := RunPa11y(analysis.URL, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running pa11y for %s: %v\n", analysis.URL, err)
				w.service.UpdateResult(analysis.ID, StatusFailed, nil)
				continue
			}

			w.service.UpdateResult(analysis.ID, StatusCompleted, result)
		}
	}()
}
