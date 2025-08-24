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

## Deployment

This application can be deployed to [Koyeb](https://www.koyeb.com/) using the provided GitHub Actions workflow. The workflow will automatically deploy the application whenever you push changes to the `main` branch.

### Prerequisites

Before you can deploy the application, you need to have a [Koyeb](https://www.koyeb.com/) account.

### Setup

1.  **Create a Koyeb API Token:**
    Go to the [API access tokens settings](https://app.koyeb.com/settings/api) in the Koyeb control panel and generate a new token. Copy the token to your clipboard.

2.  **Add the token to GitHub Secrets:**
    In your GitHub repository, go to `Settings` > `Secrets and variables` > `Actions`. Create a new repository secret named `KOYEB_API_TOKEN` and paste the token you copied in the previous step.

### Continuous Deployment

Once you have completed the setup, the GitHub Actions workflow in `.github/workflows/koyeb.yml` will automatically deploy your application to Koyeb every time you push a change to the `main` branch. The application will be deployed with the name `pa11y-go-wrapper`.

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
