package service

import (
	"errors"
	"projek_uas/app/model"
	"projek_uas/app/repository"
	"projek_uas/helper"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	userRepo     *repository.UserRepository
	studentRepo  *repository.StudentRepository
	lecturerRepo *repository.LecturerRepository
}

func NewUserService(userRepo *repository.UserRepository, studentRepo *repository.StudentRepository, lecturerRepo *repository.LecturerRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.User, error) {
	// Check if username exists
	existing, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Get role
	role, err := s.userRepo.GetRoleByName(req.RoleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("invalid role")
	}

	// Hash password
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       role.ID,
		RoleName:     role.Name,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create student or lecturer profile
	if req.RoleName == "Mahasiswa" && req.StudentID != "" {
		student := &model.Student{
			UserID:       user.ID,
			StudentID:    req.StudentID,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
		}
		if err := s.studentRepo.Create(student); err != nil {
			return nil, err
		}
	} else if req.RoleName == "Dosen Wali" && req.LecturerID != "" {
		lecturer := &model.Lecturer{
			UserID:     user.ID,
			LecturerID: req.LecturerID,
			Department: req.Department,
		}
		if err := s.lecturerRepo.Create(lecturer); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) GetUsers(page, limit int) ([]*model.User, *model.Pagination, error) {
	offset := (page - 1) * limit
	users, total, err := s.userRepo.GetAll(limit, offset)
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

func (s *UserService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) UpdateUser(id string, req *model.UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.Update(id, req)
}

func (s *UserService) DeleteUser(id string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(id)
}

func (s *UserService) HandleCreateUser(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := s.CreateUser(&req)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User created successfully", user)
}

func (s *UserService) HandleGetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	users, pagination, err := s.GetUsers(page, limit)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.PaginatedResponse(c, users, *pagination)
}

func (s *UserService) HandleGetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.GetUserByID(id)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return helper.SuccessResponse(c, "User retrieved", user)
}

func (s *UserService) HandleUpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := s.UpdateUser(id, &req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User updated successfully", nil)
}

func (s *UserService) HandleDeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.DeleteUser(id); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "User deleted successfully", nil)
}
