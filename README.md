# pa11y-go-wrapper

An API server that wraps the `pa11y` accessibility analyzer. It allows you to perform direct analysis or queue pages for analysis.

## Installation

### Prerequisites

Before you can use this tool, you need to have the following installed on your system:

*   **Go**: This tool is written in Go, so you'll need to have the Go toolchain installed. You can find installation instructions on the [official Go website](https://golang.org/doc/install).
*   **Node.js and npm**: `pa11y` is a Node.js-based tool, so you'll need to have Node.js and npm installed. You can download them from the [official Node.js website](https://nodejs.org/).
*   **pa11y**: The `pa11y` command-line tool must be installed globally. You can install it using npm:
    ```bash
    npm install -g pa11y
    ```

### Building the server

Once you have the prerequisites installed, you can build the server from source:

```bash
go build -o pa11y-go-server cmd/server/main.go
```

This will create an executable file named `pa11y-go-server` in the current directory.

## Usage

To start the server, run the following command:

```bash
./pa11y-go-server
```

The server will start on port 8080.

## API

The server exposes the following API endpoints:

### `POST /api/analyze`

Performs a direct (synchronous) analysis of a URL.

**Request Body:**

```json
{
  "url": "https://example.com",
  "runner": "htmlcs"
}
```

*   `url` (string, required): The URL to analyze.
*   `runner` (string, optional): The test runner to use (e.g., `htmlcs`, `axe`). Defaults to `htmlcs`.

**Response:**

The response will be the JSON output from `pa11y`.

### `POST /api/queue`

Adds a URL to the analysis queue.

**Request Body:**

```json
{
  "url": "https://example.com"
}
```

*   `url` (string, required): The URL to add to the queue.

**Response:**

The response will be a JSON object representing the newly created analysis task.

```json
{
  "id": "1662556965142624000",
  "url": "https://example.com",
  "status": "pending",
  "createdAt": "2022-09-07T13:22:45.142624Z",
  "updatedAt": "2022-09-07T13:22:45.142624Z"
}
```

### `GET /api/queue`

Lists all analysis tasks and their statuses.

**Response:**

The response will be a JSON array of analysis tasks.

### `GET /api/queue/:id`

Retrieves the details and analysis result of a specific task.

**Response:**

The response will be a JSON object representing the analysis task. If the analysis is complete, the `result` field will contain the `pa11y` output.
