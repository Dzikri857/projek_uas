# Setup Instructions

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- MongoDB 4.4 or higher

## Installation Steps

### 1. Clone or Extract Project

```bash
cd projek_uas
```

### 2. Install Dependencies

Run the following command to download all required Go modules:

```bash
go mod download
```

Or if you prefer to tidy up and download:

```bash
go mod tidy
```

### 3. Setup Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Edit `.env` file with your database credentials:

```env
# Server Configuration
SERVER_PORT=8080

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=your_username
POSTGRES_PASSWORD=your_password
POSTGRES_DB=student_achievement

# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=student_achievement

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRE_HOURS=24

# Log Configuration
LOG_FILE=logs/app.log
LOG_MAX_SIZE=10
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=30
```

### 4. Setup Databases

#### PostgreSQL

Create the database:

```sql
CREATE DATABASE student_achievement;
```

The application will automatically create the required tables and seed initial data on first run.

#### MongoDB

No manual setup required. The application will connect to MongoDB and create collections as needed.

### 5. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### 6. Verify Installation

Check health status:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "success",
  "postgres": "connected",
  "mongodb": "connected"
}
```

## Default Roles

The system automatically creates three roles:

1. **Admin** - Full system access
2. **Mahasiswa** - Can create and manage own achievements
3. **Dosen Wali** - Can verify student achievements

## Troubleshooting

### Import Cycle Error

If you encounter an import cycle error, make sure all imports are correct and there are no circular dependencies between packages.

### Missing go.sum Entries

Run:

```bash
go mod tidy
```

This will update `go.sum` with all required dependencies.

### Database Connection Issues

- Verify PostgreSQL is running: `pg_isready`
- Verify MongoDB is running: `mongosh --eval "db.version()"`
- Check your `.env` file has correct credentials
- Ensure databases are created and accessible

### Port Already in Use

If port 8080 is already in use, change `SERVER_PORT` in `.env` file.

## Development

### Project Structure

```
projek_uas/
├── app/
│   ├── model/          # Data structures
│   ├── repository/     # Database queries
│   └── service/        # Business logic
├── config/
│   ├── app.go          # App initialization
│   ├── env.go          # Environment loader
│   └── logger.go       # Logging configuration
├── database/           # Database connections
├── helper/             # Utility functions
├── logs/               # Log files
├── middleware/         # HTTP middlewares
├── route/              # Route definitions
├── .env                # Environment variables
├── .env.example        # Example environment file
└── main.go             # Entry point
```

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o student-achievement-system main.go
```

Then run:

```bash
./student-achievement-system
