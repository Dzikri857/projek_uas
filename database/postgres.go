package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var PostgresDB *sql.DB

func ConnectPostgres(host, port, user, password, dbname string) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	PostgresDB = db
	log.Println("Connected to PostgreSQL")

	// Initialize tables
	if err := initializeTables(); err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	return nil
}

func initializeTables() error {
	schema := `
	-- Create roles table
	CREATE TABLE IF NOT EXISTS roles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(50) UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create permissions table
	CREATE TABLE IF NOT EXISTS permissions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) UNIQUE NOT NULL,
		resource VARCHAR(50) NOT NULL,
		action VARCHAR(50) NOT NULL,
		description TEXT
	);

	-- Create role_permissions table
	CREATE TABLE IF NOT EXISTS role_permissions (
		role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
		permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
		PRIMARY KEY (role_id, permission_id)
	);

	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(100) NOT NULL,
		role_id UUID REFERENCES roles(id),
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create lecturers table
	CREATE TABLE IF NOT EXISTS lecturers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		lecturer_id VARCHAR(20) UNIQUE NOT NULL,
		department VARCHAR(100),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create students table
	CREATE TABLE IF NOT EXISTS students (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		student_id VARCHAR(20) UNIQUE NOT NULL,
		program_study VARCHAR(100),
		academic_year VARCHAR(10),
		advisor_id UUID REFERENCES lecturers(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create achievement_references table
	CREATE TABLE IF NOT EXISTS achievement_references (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		student_id UUID REFERENCES students(id) ON DELETE CASCADE,
		mongo_achievement_id VARCHAR(24) NOT NULL,
		status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'verified', 'rejected')),
		submitted_at TIMESTAMP,
		verified_at TIMESTAMP,
		verified_by UUID REFERENCES users(id),
		rejection_note TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_achievement_student ON achievement_references(student_id);
	CREATE INDEX IF NOT EXISTS idx_achievement_status ON achievement_references(status);
	`

	_, err := PostgresDB.Exec(schema)
	if err != nil {
		return err
	}

	// Insert default roles and permissions
	if err := seedDefaultData(); err != nil {
		return err
	}

	return nil
}

func seedDefaultData() error {
	// Check if roles already exist
	var count int
	err := PostgresDB.QueryRow("SELECT COUNT(*) FROM roles").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Insert roles
	roles := []struct {
		name        string
		description string
	}{
		{"Admin", "System administrator with full access"},
		{"Mahasiswa", "Student who can submit achievements"},
		{"Dosen Wali", "Academic advisor who verifies student achievements"},
	}

	roleIDs := make(map[string]string)
	for _, role := range roles {
		var id string
		err := PostgresDB.QueryRow(
			"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
			role.name, role.description,
		).Scan(&id)
		if err != nil {
			return err
		}
		roleIDs[role.name] = id
	}

	// Insert permissions
	permissions := []struct {
		name        string
		resource    string
		action      string
		description string
	}{
		{"achievement:create", "achievement", "create", "Create new achievement"},
		{"achievement:read", "achievement", "read", "Read achievement details"},
		{"achievement:update", "achievement", "update", "Update achievement"},
		{"achievement:delete", "achievement", "delete", "Delete achievement"},
		{"achievement:verify", "achievement", "verify", "Verify student achievement"},
		{"user:manage", "user", "manage", "Manage users"},
		{"report:view", "report", "view", "View reports"},
	}

	permissionIDs := make(map[string]string)
	for _, perm := range permissions {
		var id string
		err := PostgresDB.QueryRow(
			"INSERT INTO permissions (name, resource, action, description) VALUES ($1, $2, $3, $4) RETURNING id",
			perm.name, perm.resource, perm.action, perm.description,
		).Scan(&id)
		if err != nil {
			return err
		}
		permissionIDs[perm.name] = id
	}

	// Assign permissions to roles
	rolePermissions := map[string][]string{
		"Admin": {
			"achievement:create", "achievement:read", "achievement:update",
			"achievement:delete", "achievement:verify", "user:manage", "report:view",
		},
		"Mahasiswa": {
			"achievement:create", "achievement:read", "achievement:update", "achievement:delete",
		},
		"Dosen Wali": {
			"achievement:read", "achievement:verify", "report:view",
		},
	}

	for roleName, perms := range rolePermissions {
		roleID := roleIDs[roleName]
		for _, permName := range perms {
			permID := permissionIDs[permName]
			_, err := PostgresDB.Exec(
				"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)",
				roleID, permID,
			)
			if err != nil {
				return err
			}
		}
	}

	log.Println("Default roles and permissions seeded")
	return nil
}

func ClosePostgres() {
	if PostgresDB != nil {
		PostgresDB.Close()
	}
}
