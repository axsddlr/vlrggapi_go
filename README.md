# vlrggapi

vlrggapi is an open-source REST API written in Go (Golang) that scrapes and serves data from [vlr.gg](https://www.vlr.gg), a popular site for Valorant esports news, stats, rankings, and match results. The API is built using [Fiber](https://gofiber.io/), a fast, Express-inspired web framework for Go.

## Features

- **/vlr/news**: Get the latest Valorant esports news articles.
- **/vlr/stats**: Retrieve player statistics, filterable by region and timespan.
- **/vlr/rankings**: Get team rankings for different regions.
- **/vlr/match**: Fetch recent match results with detailed info.
- **/vlr/live**: Get live match scores and details.
- **/vlr/events**: Get upcoming and completed Valorant events (with `upcoming` and `completed` query params).
- **/vlr/health**: Health check for the API and upstream sources.

### Improvements

- **Structured Logging:** Uses [zap](https://github.com/uber-go/zap) for structured, production-grade logging.
- **In-memory Caching:** GET requests are cached in-memory for 30 seconds to reduce load and improve response times. Cache is per-endpoint+query.
- **Extensible Scraper Registry:** New scrapers can be added by implementing the `Scraper` interface and registering with `RegisterScraper`, making it easy to add new endpoints or data sources.

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
- (Optional) [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)

### Clone and Build (Native Go)

```bash
git clone https://github.com/yourusername/vlrggapi.git
cd vlrggapi
go mod tidy
go build -o vlrggapi ./cmd
```

### Run (Native Go)

```bash
go run cmd/main.go
```

The server will start on `http://localhost:3001` by default.

---

### Run with Docker

Build and run the container:

```bash
docker build -t vlrggapi .
docker run -p 3001:3001 --name vlrggapi --rm vlrggapi
```

Or use Docker Compose:

```bash
docker-compose up --build
```

The API will be available at [http://localhost:3001](http://localhost:3001).

---

### Notes on Performance & Logging

- All GET endpoints are cached in-memory for 30 seconds by default. You can adjust the cache TTL in `cmd/main.go`.
- All errors and important events are logged using zap for easier debugging and monitoring.

## API Documentation

Interactive Swagger UI is available at [http://localhost:3001/swagger/index.html](http://localhost:3001/swagger/index.html) after running the server.

---

## Usage

Once running, you can access the API endpoints using any HTTP client (browser, curl, Postman, etc).

- API documentation (placeholder): [http://localhost:3001/docs](http://localhost:3001/docs)
- Example: [http://localhost:3001/vlr/news](http://localhost:3001/vlr/news)

---

## API Endpoints

### `/vlr/news`
- **GET**: Returns a list of recent Valorant news articles.

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

### `/vlr/live`
- **GET**: Returns live match scores and details.

### `/vlr/events`
- **GET**: Returns Valorant events.
- **Query Parameters:**
  - `upcoming` (optional): Show only upcoming events (e.g. `/vlr/events?upcoming` or `/vlr/events?upcoming=true`)
  - `completed` (optional): Show only completed events (e.g. `/vlr/events?completed` or `/vlr/events?completed=true`)
  - If neither is set, both are shown.

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
│   │   ├── matches.go    # Match results, live scores scraping
│   │   ├── rankings.go   # Rankings scraping
│   │   ├── stats.go      # Stats scraping
│   │   └── health.go     # Health check
│   └── utils/
│       └── utils.go      # Shared headers, region map, etc.
├── docs/                 # Swagger/OpenAPI generated docs
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
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
