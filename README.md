# Echo App - WebSocket Chat + Auth (Go, PostgreSQL, Redis, Docker)

Echo App adalah aplikasi backend berbasis **Go (Golang)** yang menggunakan framework **Echo v4**, **JWT Authentication**, **WebSocket**, **PostgreSQL**, dan **Redis**.  
Proyek ini dirancang sebagai backend modern untuk menangani API, autentikasi, serta komunikasi real-time.

---

## 🚀 Tech Stack

- [Go 1.25.1](https://go.dev/)  
- [Echo v4](https://echo.labstack.com/) – Web framework  
- [Gorilla WebSocket](https://github.com/gorilla/websocket) – Real-time communication  
- [PostgreSQL](https://www.postgresql.org/) + [GORM](https://gorm.io/) – Database ORM  
- [Redis](https://redis.io/) – Caching & Pub/Sub  
- [JWT](https://github.com/golang-jwt/jwt) – Authentication  
- [godotenv](https://github.com/joho/godotenv) – Environment variable loader  

---

## 📦 Installation

### 1. Clone repository
```bash
git clone https://github.com/username/echo-app.git
cd echo-app
```

### 2. Buat file `.env`
Contoh konfigurasi:
```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=yourdbname

REDIS_ADDR=localhost:6379
JWT_SECRET=supersecret
```

### 3. Jalankan PostgreSQL & Redis (opsional pakai Docker)
```bash
# PostgreSQL
docker run --name postgres -e POSTGRES_PASSWORD=yourpassword -p 5432:5432 -d postgres:15

# Redis
docker run --name redis -p 6379:6379 -d redis:7
```

### 4. Jalankan aplikasi
```bash
go run main.go
```

---

## 📡 API Features

- 🔑 **Authentication**  
  - Login dengan JWT  
  - Middleware proteksi endpoint  

- 🗄 **Database Integration**  
  - PostgreSQL via GORM  
  - Migration otomatis  

- 💬 **WebSocket Support**  
  - Real-time chat/messaging  
  - Redis Pub/Sub untuk scale-out  

---

## 🧪 Testing

Gunakan [Postman](https://www.postman.com/) atau [cURL](https://curl.se/) untuk menguji endpoint.  
Contoh cek server:

```bash
curl http://localhost:8080/
```

---

## 📜 License

Proyek ini dirilis di bawah lisensi [MIT License](LICENSE).  
