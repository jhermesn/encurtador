# Encurtador

A fast, self-hosted URL shortener written in Go with a React front-end hosted on GitHub Pages.
The Go service runs as a single Docker container (MySQL and Redis are expected to already be available on the host).

## Features

- **Custom slugs** -> choose your own readable short code (5–50 chars) or let the system generate an 8-character random one
- **TTL options** -> links expire after 1 hour, 1 day, 1 week, 1 month, or 1 year
- **Password protection** -> optional bcrypt-hashed password gates access to the redirect
- **Management tokens** -> each link gets a one-time management token; use it to expire the URL early
- **Cache-aside** -> Redis sits in front of MySQL; the redirect hot path almost never hits the database
- **Rate limiting** -> redirect and password-unlock endpoints are capped at 60 requests/minute per IP
- **No account required** -> open panel, anyone can create a link

---

## Architecture

```mermaid
flowchart TD
    Browser -->|"loads SPA"| GHPages["GitHub Pages"]
    Browser -->|"API calls | Origin: jhermesn.dev"| NPM["Nginx Proxy Manager"]
    NPM --> API["Go/Gin API :8080 | CORS: allow jhermesn.dev"]

    API --> RL["Rate Limiter\n60 req/min per IP"]
    RL --> Redirect["GET /:slug"]
    RL --> Unlock["POST /api/v1/urls/:slug/unlock"]
    API --> Create["POST /api/v1/urls"]
    API --> Check["GET /api/v1/urls/check/:slug"]
    API --> Expire["POST /api/v1/urls/:slug/expire"]
    Redirect --> Redis["Redis (cache-aside)"]
    Unlock --> Redis
    Redis -->|"cache miss"| MySQL["MySQL"]
    Create --> MySQL
    Create -->|"pre-warm"| Redis
    Expire --> MySQL
    Expire -->|"invalidate"| Redis
    MySQL --> Cleanup["Cleanup Goroutine, hourly DELETE expired rows"]
```

I use NGINX Proxy Manager due to my infrastructure security, but you can host the api anywhere you want.

The **React SPA is hosted on GitHub Pages** at `https://jhermesn.dev/encurtador/` and talks to the Go API via CORS.
MySQL and Redis run locally on the host and are reachable via `network_mode: host` inside Docker.

### Logs → Loki using my [homelab infrastructure](https://github.com/jhermesn/homelab-infra?tab=readme-ov-file)

Promtail is configured to scrape containers with the label `logging=promtail` and ship their stdout to Loki.
`docker-compose.yml` sets:

- `logging=promtail`
- `service=encurtador`

The server logs as JSON via `slog` and emits a `timestamp` field so Promtail can parse and label logs consistently.

---

## Database Model

```mermaid
erDiagram
    urls {
        BIGINT_UNSIGNED id PK
        VARCHAR_100 slug UK "unique short code"
        TEXT target_url "destination URL"
        VARCHAR_60 password_hash "NULL = no password (bcrypt)"
        CHAR_64 manage_token_hash "SHA-256 of the management token"
        TIMESTAMP expires_at "indexed for cleanup"
        TIMESTAMP created_at
    }
```

The schema is created automatically on first startup via an idempotent `CREATE TABLE IF NOT EXISTS`.

### Query performance

| Query | Index used | Notes |
|---|---|---|
| `SELECT ... WHERE slug = ? AND expires_at > NOW()` | UNIQUE on `slug` | B-tree point lookup, dominant hot path, almost always served by Redis |
| `SELECT EXISTS(SELECT 1 WHERE slug = ?)` | UNIQUE on `slug` | availability check |
| `INSERT INTO urls ...` | UNIQUE on `slug` | single-row write |
| `UPDATE ... SET expires_at = NOW() WHERE slug = ?` | UNIQUE on `slug` | early expire, rare |
| `DELETE WHERE expires_at < NOW()` | INDEX on `expires_at` | hourly batch cleanup |

---

## Request Flows

### Redirect (hot path)

```mermaid
sequenceDiagram
    participant Browser
    participant Gin
    participant Redis
    participant MySQL

    Browser->>Gin: GET /:slug
    Gin->>Redis: GET url:{slug}
    alt cache hit - not protected
        Redis-->>Gin: payload
        Gin-->>Browser: 301 to target_url
    else cache hit - password protected
        Redis-->>Gin: payload
        Gin-->>Browser: 302 to jhermesn.dev/encurtador/gate/:slug
        Browser->>Gin: POST /api/v1/urls/:slug/unlock {password}
        Gin->>Redis: GET url:{slug}
        Redis-->>Gin: payload with bcrypt hash
        Gin->>Gin: bcrypt.Compare
        Gin-->>Browser: 200 {target_url} or 401
    else cache miss
        Redis-->>Gin: nil
        Gin->>MySQL: SELECT WHERE slug=? AND expires_at>NOW()
        alt row found
            MySQL-->>Gin: row
            Gin->>Redis: SET url:{slug} EX remaining_ttl
            Gin-->>Browser: 301 to target_url or 302 to gate
        else not found or expired
            MySQL-->>Gin: no rows
            Gin-->>Browser: 404
        end
    end
```

### URL Creation

```mermaid
sequenceDiagram
    participant Browser
    participant Gin
    participant MySQL
    participant Redis

    Browser->>Gin: POST /api/v1/urls {target_url, slug?, ttl, password?}
    Gin->>Gin: validate, resolve slug collisions
    Gin->>Gin: bcrypt.Hash(password) if set
    Gin->>Gin: crypto/rand → manage_token (32 chars base62)
    Gin->>Gin: sha256(manage_token) → manage_token_hash
    Gin->>MySQL: INSERT INTO urls
    Gin->>Redis: SET url:{slug} EX ttl (pre-warm)
    Gin-->>Browser: 201 {slug, short_url, expires_at, protected, manage_token}
    Note over Browser: manage_token shown ONCE - store it safely
```

### Early Expire

```mermaid
sequenceDiagram
    participant Browser
    participant Gin
    participant MySQL
    participant Redis

    Browser->>Gin: POST /api/v1/urls/:slug/expire {manage_token}
    Gin->>Gin: sha256(manage_token) → hash
    Gin->>MySQL: UPDATE SET expires_at=NOW() WHERE slug=? AND manage_token_hash=?
    alt 1 row affected
        Gin->>Redis: DEL url:{slug}
        Gin-->>Browser: 200 {message}
    else 0 rows
        Gin-->>Browser: 401
    end
```

---

## API Reference

| Method | Path | Body | Response |
|---|---|---|---|
| `POST` | `/api/v1/urls` | `{target_url, slug?, ttl, password?}` | `201 {slug, short_url, expires_at, protected, manage_token}` |
| `GET`  | `/api/v1/urls/check/:slug` | - | `200 {available, suggestion?}` |
| `GET`  | `/:slug` | - | `301` redirect, `302` to frontend gate page, or `404` |
| `POST` | `/api/v1/urls/:slug/unlock` | `{password}` | `200 {target_url}` or `401` |
| `POST` | `/api/v1/urls/:slug/expire` | `{manage_token}` | `200` or `401` |

**TTL values:** `1h` · `24h` · `168h` · `720h` · `8760h`

Rate limiting (60 req/min per IP, shared counter across redirect + unlock) applies to `GET /:slug` and `POST /api/v1/urls/:slug/unlock`. Exceeding the limit returns `429`.

---

## Environment Variables

### Backend (`api/.env`)

| Variable | Required | Default | Description |
|---|---|---|---|
| `MYSQL_DSN` | ✓ | - | MySQL DSN, e.g. `user:pass@tcp(127.0.0.1:3306)/encurtador?parseTime=true` |
| `REDIS_ADDR` | ✓ | - | Redis address, e.g. `127.0.0.1:6379` |
| `BASE_URL` | ✓ | - | Public base URL without trailing slash, e.g. `https://encurtador.jhermesn.dev` |
| `APP_PORT` | - | `8080` | Port to listen on |
| `CORS_ALLOWED_ORIGIN` | - | `https://jhermesn.dev` | Origin allowed to make cross-origin requests |
| `FRONTEND_URL` | - | `https://jhermesn.dev/encurtador` | Frontend base path; used when redirecting to the password gate |

### Frontend (`web/.env`)

| Variable | Required | Description |
|---|---|---|
| `VITE_API_URL` | ✓ | Backend API base URL, e.g. `https://encurtador.jhermesn.dev` |

---

## Some Notes

- **Link passwords** use bcrypt. The hash is stored in MySQL and cached in Redis; the plain-text password is never persisted.
- **Management tokens** are 32-character cryptographically random base62 strings generated with rejection sampling to eliminate modulo bias. Only the SHA-256 hash is stored - the plain token is returned once at creation time.
- **Auto-generated slugs** use `crypto/rand` with 8 base62 characters (~218 trillion combinations), making enumeration impractical.
- **Rate limiting** (60 req/min per IP, shared counter across redirect + unlock) stops real-time brute-force attacks.
