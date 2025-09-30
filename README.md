# Echo App - WebSocket Chat + Auth (Go, PostgreSQL, Redis, Docker)

Project ini adalah aplikasi **monolith** berbasis [Echo Framework](https://echo.labstack.com/) (Go),  
dengan fitur **Auth (JWT)**, **Profile & Roles**, dan **Chat Realtime WebSocket**.  
Database menggunakan **PostgreSQL**, Redis dipakai untuk **Pub/Sub** agar siap scale out.  
Seluruh service berjalan dengan **Docker Compose**.

---

## 🚀 Fitur
- Register / Login dengan JWT
- CRUD Roles & Profiles
- Ganti email & password
- Forgot & reset password
- Chat Realtime via WebSocket
- Redis Pub/Sub untuk broadcast pesan antar instance

---

## 📦 Requirements
- [Go](https://go.dev/) >= 1.25 (opsional, hanya kalau mau run lokal tanpa docker)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

---

## 🔧 Setup & Jalankan

1. **Clone repository**
   ```bash
   git clone https://github.com/username/echo-app.git
   cd echo-app
