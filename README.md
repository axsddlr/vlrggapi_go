# vlrggapi

vlrggapi is an open-source REST API written in Go (Golang) that scrapes and serves data from [vlr.gg](https://www.vlr.gg), a popular site for Valorant esports news, stats, rankings, and match results. The API is built using [Fiber](https://gofiber.io/), a fast, Express-inspired web framework for Go.

## Features

- **/vlr/news**: Get the latest Valorant esports news articles.
- **/vlr/stats**: Retrieve player statistics, filterable by region and timespan.
- **/vlr/rankings**: Get team rankings for different regions.
- **/vlr/match**: Fetch recent match results with detailed info.
- **/vlr/health**: Health check for the API and upstream sources.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Environment Variables](#environment-variables)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

### Prerequisites

- Go 1.24+ installed ([download here](https://go.dev/dl/))
- Git

### Clone and Build

```bash
git clone https://github.com/yourusername/vlrggapi.git
cd vlrggapi
go mod tidy
go build -o vlrggapi ./cmd
```

### Run

```bash
go run cmd/main.go
```

The server will start on `http://localhost:3001` by default.

---

## Usage

Once running, you can access the API endpoints using any HTTP client (browser, curl, Postman, etc).

- API documentation (placeholder): [http://localhost:3001/docs](http://localhost:3001/docs)
- Example: [http://localhost:3001/vlr/news](http://localhost:3001/vlr/news)

---

## API Endpoints

### `/vlr/news`

- **GET**: Returns a list of recent Valorant news articles.
- **Response Example:**
  ```json
  {
    "data": {
      "status": 200,
      "segments": [
        {
          "title": "FUT announces return of cNed, arrival of Johnta",
          "description": "The former world champion is back after a short break.",
          "date": "July 14, 2025",
          "author": "jenopelle",
          "url_path": "https://vlr.gg/516910/fut-announces-return-of-cned-arrival-of-johnta"
        }
      ]
    }
  }
  ```

### `/vlr/stats`

- **GET**: Returns player statistics.
- **Query Parameters:**
  - `region` (required): e.g. `na`, `eu`, `ap`, etc.
  - `timespan` (optional): e.g. `all`, `30` (days)
- **Example:** `/vlr/stats?region=na&timespan=30`

### `/vlr/rankings`

- **GET**: Returns team rankings for a region.
- **Query Parameters:**
  - `region` (required): e.g. `na`, `eu`, `ap`, etc.
- **Example:** `/vlr/rankings?region=eu`

### `/vlr/match`

- **GET**: Returns recent match results.
- **Query Parameters:**
  - `num_pages` (optional): Number of pages to fetch (default: 1)
  - `from_page`, `to_page` (optional): Page range
  - `max_retries` (optional): Retry attempts per page (default: 3)
  - `request_delay` (optional): Delay between requests in seconds (default: 1.0)
  - `timeout` (optional): HTTP timeout in seconds (default: 30)

### `/vlr/health`

- **GET**: Returns health status of the API and upstream sources.

---

## Environment Variables

- `PORT`: The port to run the server on (default: `3001`).

---

## Project Structure

```
vlrggapi/
├── cmd/
│   └── main.go           # Application entrypoint
├── internal/
│   ├── router/
│   │   └── vlr_router.go # Route registration
│   ├── scrapers/
│   │   ├── news.go       # News scraping logic
│   │   ├── matches.go    # Match results scraping
│   │   ├── rankings.go   # Rankings scraping
│   │   ├── stats.go      # Stats scraping
│   │   └── health.go     # Health check
│   └── utils/
│       └── utils.go      # Shared headers, region map, etc.
├── go.mod
├── go.sum
└── README.md
```

---

## Contributing

Pull requests and issues are welcome! Please open an issue to discuss your ideas or report bugs.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'Add new feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a pull request

---

## License

This project is licensed under the MIT License.
