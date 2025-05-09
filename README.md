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

## 🧾 License

MIT

```

Let me know if you want to include instructions for Docker, Swagger docs, or unit testing.
```
