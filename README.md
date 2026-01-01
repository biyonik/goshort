# GoShort

A high-performance URL shortener service built with Go and Angular.

## Features

- **URL Shortening**: Convert long URLs to short, shareable links
- **Fast Redirects**: Sub-10ms redirect latency with Redis caching
- **Click Analytics**: Track clicks, referrers, and geographic data
- **Custom URLs**: Create branded short links (premium)
- **Rate Limiting**: Token bucket algorithm for abuse prevention
- **URL Expiration**: Auto-expire links after specified duration

## Tech Stack

### Backend (`/api`)
- **Go** with Gin framework
- **PostgreSQL** - URL and user storage
- **Redis** - Caching and rate limiting
- **ClickHouse** - Analytics and time-series data

### Frontend (`/ui`)
- **Angular 17+** with standalone components
- **Tailwind CSS** for styling
- **ngx-charts** for analytics visualization

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Gin API   │────▶│    Redis    │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ Postgres │ │ ClickHouse│ │  Worker  │
        └──────────┘ └──────────┘ └──────────┘
```

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose

### Quick Start

```bash
# Clone the repository
git clone https://github.com/ahmethakanbesel/goshort.git
cd goshort

# Start infrastructure
cd api
docker-compose up -d

# Run the API
go run cmd/api/main.go

# In another terminal, start the UI
cd ui
npm install
ng serve
```

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/urls` | Create short URL |
| GET | `/:code` | Redirect to long URL |
| GET | `/api/urls/:id/analytics` | Get click analytics |
| GET | `/health` | Health check |

## Configuration

Copy `.env.example` to `.env` and configure:

```env
SERVER_PORT=8080
BASE_URL=http://localhost:8080
POSTGRES_HOST=localhost
REDIS_HOST=localhost
SHORT_CODE_LENGTH=7
```

## Project Structure

```
goshort/
├── api/                    # Go backend
│   ├── cmd/api/            # Application entrypoint
│   ├── internal/           # Private application code
│   │   ├── config/         # Configuration
│   │   ├── domain/         # Domain models
│   │   ├── handler/        # HTTP handlers
│   │   ├── repository/     # Data access layer
│   │   ├── service/        # Business logic
│   │   └── shortener/      # URL shortening logic
│   ├── pkg/                # Public libraries
│   └── migrations/         # Database migrations
│
└── ui/                     # Angular frontend
    └── src/
        ├── app/
        └── environments/
```

## License

MIT
