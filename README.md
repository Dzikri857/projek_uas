# Student Achievement Reporting System

Backend API untuk sistem pelaporan prestasi mahasiswa dengan Role-Based Access Control (RBAC).

## Fitur

- Autentikasi dan otorisasi berbasis JWT
- Role-Based Access Control (Admin, Mahasiswa, Dosen Wali)
- CRUD prestasi mahasiswa dengan field dinamis
- Workflow verifikasi prestasi
- Statistik dan pelaporan prestasi
- Dual database: PostgreSQL (RBAC) + MongoDB (prestasi dinamis)

## Tech Stack

- **Backend**: Go 1.25 + Fiber v2
- **Database**: PostgreSQL + MongoDB
- **Authentication**: JWT
- **Password Hashing**: bcrypt

## Struktur Database

### PostgreSQL
- users, roles, permissions, role_permissions
- students, lecturers
- achievement_references (tracking status)

### MongoDB
- achievements (data prestasi dengan field dinamis)

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Configure database connections in `.env`

4. Run the application:
```bash
go run main.go
```

Server akan berjalan di `http://localhost:3000`

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/profile` - Get profile
- `POST /api/v1/auth/logout` - Logout

### Users (Admin only)
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user detail
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Achievements
- `GET /api/v1/achievements` - List achievements
- `GET /api/v1/achievements/:id` - Get achievement detail
- `POST /api/v1/achievements` - Create achievement
- `PUT /api/v1/achievements/:id` - Update achievement
- `DELETE /api/v1/achievements/:id` - Delete achievement
- `POST /api/v1/achievements/:id/submit` - Submit for verification
- `POST /api/v1/achievements/:id/verify` - Verify/Reject achievement

### Reports
- `GET /api/v1/reports/statistics` - Get statistics

## Default Roles & Permissions

### Admin
- Full access ke semua fitur

### Mahasiswa
- Create, read, update, delete prestasi sendiri
- Submit prestasi untuk verifikasi

### Dosen Wali
- Read prestasi mahasiswa bimbingan
- Verify/reject prestasi

## Catatan

- Sistem ini **TIDAK menggunakan fitur notification** sesuai permintaan
- Password di-hash menggunakan bcrypt
- JWT token expires dalam 24 jam
- Refresh token expires dalam 168 jam (7 hari)
