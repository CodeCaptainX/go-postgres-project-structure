# go-postgres-project-structure
Here is your complete `README.md` as a **code file**:

```markdown
# 🛠️ Go Postgres Project

A structured Go Fiber API server with PostgreSQL, Redis, JWT authentication, and internationalization support.

---

## 📁 Project Structure

```

.
├── config           # Configuration loader
├── docs             # API documentation
├── handler          # HTTP request handlers
├── internal         # Internal logic
│   └── migrations   # DB migration scripts
├── pkg              # Reusable packages
│   ├── constants
│   ├── custom\_log
│   ├── middleware
│   ├── model
│   ├── redis
│   ├── sql
│   ├── translates   # i18n JSON files (en.json, zh.json, km.json)
│   ├── utils
│   └── validator
├── routers          # Router definitions
├── tmp              # Temp files
├── main.go          # Entry point
├── .air.toml        # Live reload config
├── Makefile
└── README.md        # You're here

````

---

## ⚙️ Environment Setup

Create your `.env` file:

```bash
cp .env.sample .env
````

### .env Configuration Example

```env
DATABASE_URL="postgresql://postgres:123456@localhost:5432/test_db_01?sslmode=disable"

API_HOST="127.0.0.1"
API_PORT="8887"

TIME_ZONE="Asia/Phnom_Penh"

# JWT
JWT_SECRET_KEY="your_secret_key"
JWT_EXPIRE=24h
JWT_ACCESS_TOKEN_EXPIRE=1h
JWT_REFRESH_TOKEN_EXPIRE=8h

# Date
DEFAULT_FORMAT_DATE="Y/m/d"
DEFAULT_FORMAT_DATE_RESPONSE="Y/m/d H:i:s AM"

# Context
PLAYER_CONTEXT="yourContext"

# Redis
REDIS_HOST="127.0.0.1"
REDIS_PORT="6379"
REDIS_PASSWORD=""
REDIS_DB_NUMBER=0
REDIS_EXPIRE=60
```

---

## 🔃 Live Reload with Air

Install Air:

```bash
go install github.com/air-verse/air@latest
```

Make sure Go binaries are in your system PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Then run the project:

```bash
air -c .air.toml
```

---

## 🗃️ Database Migration

Ensure you have a migration tool like `migrate` or `goose`.

### Migrate Up (Create Tables)

```bash
make up
```

### Migrate Down (Drop All Tables)

```bash
make down
```

Your `Makefile` should include:

```makefile
up:
	migrate -path internal/migrations -database $$DATABASE_URL up

down:
	migrate -path internal/migrations -database $$DATABASE_URL down
```

---

## 🌍 Internationalization (i18n)

Language translation files are stored in:

```
pkg/translates/
├── en.json
├── zh.json
└── km.json
```

Supports English, Khmer, and Chinese via Fiber i18n middleware.

---

## 🚀 Sample API

* `GET /` → returns localized "welcome"
* `GET /:name` → returns "welcomeWithName" with `:name` replaced

---

## 📦 Dependencies

* [Fiber](https://github.com/gofiber/fiber)
* [go-i18n](https://github.com/nicksnyder/go-i18n)
* [PostgreSQL](https://www.postgresql.org/)
* [Redis](https://redis.io/)
* [Air](https://github.com/air-verse/air) for hot reload
* `make` for build tasks

---

# Goose Setup Guide
This project uses [Goose](https://github.com/pressly/goose) for managing PostgreSQL database migrations.
## Install Goose
````markdown

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
````

Make sure your `$GOPATH/bin` is in your `$PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

To make it permanent, add the line above to your shell config file (`~/.bashrc`, `~/.zshrc`, etc.).

## Common Commands

### Run all migrations

```bash
goose -dir migrations postgres "postgresql://postgres:123456@localhost:5432/test_db_01?sslmode=disable" up
```

### Rollback the last migration

```bash
goose -dir migrations postgres "postgresql://postgres:123456@localhost:5432/test_db_01?sslmode=disable" down
```

### Create a new migration

```bash
goose -dir migrations create your_migration_name sql
```

This creates a new timestamped `.sql` file inside the `migrations/` directory.

# Redis Installation Guide
This project may require Redis for caching, session storage, or other purposes.

## Install Redis on Ubuntu

### 1. Update package index
````markdown


```bash
sudo apt update
````

### 2. Install Redis server

```bash
sudo apt install redis-server -y
```

### 3. Enable and start Redis

```bash
sudo systemctl enable redis
sudo systemctl start redis
```

### 4. Check Redis status

```bash
sudo systemctl status redis
```

You should see `active (running)` in the output.

### 5. Test Redis

```bash
redis-cli ping
```

Expected output:

```
PONG
```

## Configuration (Optional)

If you want Redis to run as a background service:

1. Open the Redis config file:

```bash
sudo nano /etc/redis/redis.conf
```

2. Find the line:

```
supervised no
```

3. Change it to:

```
supervised systemd
```

4. Save and restart Redis:

```bash
sudo systemctl restart redis
```

## Uninstall Redis (if needed)

```bash
sudo apt remove redis-server -y
```

## Useful Commands

```bash
redis-cli            # Open Redis command-line interface
redis-cli ping       # Check if Redis is working
redis-cli flushall   # Clear all keys from all databases (⚠️ use with caution)
```


