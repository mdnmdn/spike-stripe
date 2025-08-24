# pa11y-go-wrapper

A command-line tool that wraps the `pa11y` accessibility analyzer and provides the output in JSON format.

## Installation

### Prerequisites

Before you can use this tool, you need to have the following installed on your system:

*   **Go**: This tool is written in Go, so you'll need to have the Go toolchain installed. You can find installation instructions on the [official Go website](https://golang.org/doc/install).
*   **Node.js and npm**: `pa11y` is a Node.js-based tool, so you'll need to have Node.js and npm installed. You can download them from the [official Node.js website](https://nodejs.org/).
*   **pa11y**: The `pa11y` command-line tool must be installed globally. You can install it using npm:
    ```bash
    npm install -g pa11y
    ```

### Building the wrapper

Once you have the prerequisites installed, you can build the wrapper from source:

```bash
go build main.go -o pa11y-go-wrapper
```

This will create an executable file named `pa11y-go-wrapper` in the current directory. You can move this file to a directory in your system's PATH to make it accessible from anywhere.

## Usage

### Checking the installation

To verify that all dependencies are installed correctly, you can use the `--check-install` flag:

```bash
./pa11y-go-wrapper --check-install
```

If everything is set up correctly, you will see the following output:

```
pa11y is installed and ready to use.
```

If `pa11y` is not installed or not in your PATH, you will see an error message.

### Analyzing a web page

To analyze a web page for accessibility issues, simply provide the URL as a command-line argument:

```bash
./pa11y-go-wrapper https://example.com
```

The tool will run `pa11y` and print a JSON report of the accessibility issues found on the page.

## Dependencies

*   [Go](https://golang.org/)
*   [Node.js](https://nodejs.org/)
*   [pa11y](https://github.com/pa11y/pa11y)
