package repository

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"projek_uas/app/model"
	"projek_uas/database"
	"projek_uas/helper"

	"github.com/gofiber/fiber/v2"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return database.PostgresDB.QueryRow(
		query,
		user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active,
		       u.created_at, u.updated_at, u.deleted_at, r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 AND u.deleted_at IS NULL
	`
	err := database.PostgresDB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.RoleName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active,
		       u.created_at, u.updated_at, u.deleted_at, r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	err := database.PostgresDB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleID,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.RoleName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) GetUserPermissions(userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE u.id = $1
	`
	rows, err := database.PostgresDB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, err
		}
		permissions = append(permissions, perm)
	}
	return permissions, nil
}

func (r *UserRepository) GetAll(limit, offset int) ([]*model.User, int64, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active,
		       u.created_at, u.updated_at, u.deleted_at, r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := database.PostgresDB.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FullName, &user.RoleID,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt, &user.RoleName,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	var total int64
	err = database.PostgresDB.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&total)
	return users, total, err
}

func (r *UserRepository) Update(id string, req *model.UpdateUserRequest) error {
	query := `
		UPDATE users
		SET email = COALESCE(NULLIF($1, ''), email),
		    full_name = COALESCE(NULLIF($2, ''), full_name),
		    is_active = COALESCE($3, is_active),
		    updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := database.PostgresDB.Exec(query, req.Email, req.FullName, req.IsActive, time.Now(), id)
	return err
}

func (r *UserRepository) SoftDelete(id string) error {
	query := `
		UPDATE users
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := database.PostgresDB.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found or already deleted")
	}

	return nil
}

func (r *UserRepository) HardDelete(id string) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := database.PostgresDB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) Restore(id string) error {
	query := `
		UPDATE users
		SET deleted_at = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NOT NULL
	`
	result, err := database.PostgresDB.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found or not deleted")
	}

	return nil
}

func (r *UserRepository) GetRoleByName(roleName string) (*model.Role, error) {
	role := &model.Role{}
	query := "SELECT id, name, description, created_at FROM roles WHERE name = $1"
	err := database.PostgresDB.QueryRow(query, roleName).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *UserRepository) HandleCreate(req *model.CreateUserRequest, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) (*model.User, error) {
	existing, err := r.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	role, err := r.GetRoleByName(req.RoleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("invalid role")
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       role.ID,
		RoleName:     role.Name,
		IsActive:     true,
	}

	if err := r.Create(user); err != nil {
		return nil, err
	}

	if req.RoleName == "Mahasiswa" && req.StudentID != "" {
		student := &model.Student{
			UserID:       user.ID,
			StudentID:    req.StudentID,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
		}
		if err := studentRepo.Create(student); err != nil {
			return nil, err
		}
	} else if req.RoleName == "Dosen Wali" && req.LecturerID != "" {
		lecturer := &model.Lecturer{
			UserID:     user.ID,
			LecturerID: req.LecturerID,
			Department: req.Department,
		}
		if err := lecturerRepo.Create(lecturer); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (r *UserRepository) HandleGetAll(page, limit int) ([]*model.User, *model.Pagination, error) {
	offset := (page - 1) * limit
	users, total, err := r.GetAll(limit, offset)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	pagination := &model.Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	return users, pagination, nil
}

func (r *UserRepository) HandleUpdate(id string, req *model.UpdateUserRequest) error {
	user, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return r.Update(id, req)
}

func (r *UserRepository) HandleDelete(id string) error {
	user, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return r.SoftDelete(id)
}

func (r *UserRepository) HandleHardDelete(id string) error {
	// Check if user exists (including soft deleted)
	query := "SELECT id FROM users WHERE id = $1"
	var userID string
	err := database.PostgresDB.QueryRow(query, id).Scan(&userID)
	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	return r.HardDelete(id)
}

func (r *UserRepository) HandleRestore(id string) error {
	return r.Restore(id)
}

func (r *UserRepository) HandleCreateHTTP(c *fiber.Ctx, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := r.HandleCreate(&req, studentRepo, lecturerRepo)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User created successfully", user)
}

func (r *UserRepository) HandleGetAllHTTP(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	users, pagination, err := r.HandleGetAll(page, limit)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.PaginatedResponse(c, users, *pagination)
}

func (r *UserRepository) HandleGetByIDHTTP(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := r.FindByID(id)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return helper.SuccessResponse(c, "User retrieved", user)
}

func (r *UserRepository) HandleUpdateHTTP(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := r.HandleUpdate(id, &req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User updated successfully", nil)
}

func (r *UserRepository) HandleDeleteHTTP(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := r.HandleDelete(id); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User deleted successfully (soft delete)", nil)
}

func (r *UserRepository) HandleHardDeleteHTTP(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := r.HandleHardDelete(id); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User permanently deleted", nil)
}

func (r *UserRepository) HandleRestoreHTTP(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := r.HandleRestore(id); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User restored successfully", nil)
}
