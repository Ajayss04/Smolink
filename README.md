# Smolink 🔗

A high-performance URL shortener API built with **Go**, **Redis**, and **MongoDB** — featuring sub-10ms redirects, click analytics, and JWT-based authentication.

## Architecture

```
Client Request
      │
      ▼
 Go Fiber API
      │
      ├──► Redis (hot path) ──► sub-5ms redirect for cached URLs
      │
      └──► MongoDB (persistent store) ──► URL metadata, user data
```

**Why two databases?**
- **Redis** sits in front as an in-memory cache. Every short URL is stored here on creation. Redirects are served entirely from RAM — no disk I/O.
- **MongoDB** is the persistent layer. Stores URL documents, user info, and survives restarts. Redis is rebuilt from Mongo on a cache miss.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.21 |
| Framework | Fiber v2 |
| Cache | Redis |
| Database | MongoDB |
| Auth | JWT (Google OAuth) |
| Containerization | Docker |

## Features

- ⚡ **Fast redirects** — short codes resolved from Redis in under 10ms
- 📊 **Click tracking** — every redirect increments a Redis counter using key namespacing (`clicks:<shortcode>`)
- 🔐 **Auth-protected routes** — JWT middleware guards all link management endpoints
- ⏳ **URL expiry** — optional TTL on short links, enforced by Redis TTL
- 🔒 **Protected URLs** — links can be secured with a secret code
- 🐳 **Dockerized** — single `docker-compose up` to run the full stack

## API Endpoints

### Public
| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/shrinkr/:shortURL` | Redirect to long URL |

### Protected (requires JWT)
| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/shrinkr/links/addURL` | Create a short URL |
| `GET` | `/shrinkr/links/mappings` | Get all links for user |
| `GET` | `/shrinkr/links/:shortURL` | Get info for one link |
| `GET` | `/shrinkr/links/:shortURL/stats` | Get click count for a link |
| `DELETE` | `/shrinkr/links/:shortURL` | Delete a link |

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- A Redis instance (local or cloud)
- A MongoDB instance (local or Atlas)

### Setup

```bash
# 1. Clone the repo
git clone https://github.com/Ajayss04/Smolink.git
cd Smolink

# 2. Copy env file and fill in your values
cp .env.example .env

# 3. Run with Docker
docker-compose up

# OR run locally
go run main.go
```

### Environment Variables

```env
FIBER_PORT=3000
REDIS_URL=redis://localhost:6379
MONGO_URI=mongodb://localhost:27017
MONGO_DB=smolink
JWT_SECRET=your_secret_here
```

## Key Design Decisions

**Redis key namespacing** — URL mappings and click counters use separate key prefixes (`<shortcode>` vs `clicks:<shortcode>`) to avoid collisions and make cache inspection easy.

**Write-through caching** — On URL creation, data is written to both MongoDB and Redis atomically. Reads always hit Redis first; MongoDB is only queried on a cache miss.

**Graceful degradation** — If Redis is unavailable, the service falls back to MongoDB. Slower, but nothing breaks.

## Project Structure

```
smolink/
├── config/       # Environment config loader
├── database/     # Redis + MongoDB connection and queries
├── handlers/     # HTTP handler functions
├── middleware/   # JWT auth guard
├── routes/       # Route definitions
├── types/        # Go structs (models, DTOs, errors)
├── main.go       # Entry point
└── Dockerfile
```