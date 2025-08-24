package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <url>")
		os.Exit(1)
	}
	url := os.Args[1]

	cmd := exec.Command("pa11y", "--reporter", "json", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// pa11y exits with code 2 if there are accessibility issues.
		// We still want to see the JSON report in that case.
		// We only exit if there's a different error.
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 2 {
				fmt.Fprintf(os.Stderr, "Error running pa11y: %v\n", err)
				fmt.Fprintf(os.Stderr, "Output: %s\n", output)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error running pa11y: %v\n", err)
			fmt.Fprintf(os.Stderr, "Output: %s\n", output)
			os.Exit(1)
		}
	}

	var result interface{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling pa11y output: %v\n", err)
		fmt.Fprintf(os.Stderr, "Output was: %s\n", output)
		os.Exit(1)
	}

	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(prettyJSON))
}
