package service

import (
	"errors"
	"projek_uas/app/model"
	"projek_uas/app/repository"
	"projek_uas/helper"
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

	return s.userRepo.SoftDelete(id)
}

func (s *UserService) HardDeleteUser(id string) error {
	return s.userRepo.HardDelete(id)
}

func (s *UserService) RestoreUser(id string) error {
	return s.userRepo.Restore(id)
}
