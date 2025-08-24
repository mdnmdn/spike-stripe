package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	checkInstall := flag.Bool("check-install", false, "Check if pa11y is installed correctly")
	flag.Parse()

	if *checkInstall {
		_, err := exec.LookPath("pa11y")
		if err != nil {
			fmt.Println("Error: pa11y is not installed or not in your PATH.")
			fmt.Println("Please install pa11y by following the instructions at https://github.com/pa11y/pa11y")
			os.Exit(1)
		}
		fmt.Println("pa11y is installed and ready to use.")
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: pa11y-go-wrapper <url>")
		fmt.Println("Use --help for more options.")
		os.Exit(1)
	}
	url := flag.Args()[0]

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
