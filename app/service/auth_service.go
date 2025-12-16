package service

import (
	"errors"
	"projek_uas/app/model"
	"projek_uas/app/repository"
	"projek_uas/config"
	"projek_uas/helper"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	if !helper.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// Get user permissions
	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return nil, err
	}
	user.Permissions = permissions

	// Generate tokens
	token, err := helper.GenerateToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	refreshToken, err := helper.GenerateRefreshToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) GetProfile(userID string) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return nil, err
	}
	user.Permissions = permissions

	return user, nil
}

func (s *AuthService) RefreshToken(oldRefreshToken string) (*model.LoginResponse, error) {
	claims, err := helper.ValidateToken(oldRefreshToken, s.cfg)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil || !user.IsActive {
		return nil, errors.New("user not found or inactive")
	}

	permissions, err := s.userRepo.GetUserPermissions(user.ID)
	if err != nil {
		return nil, err
	}
	user.Permissions = permissions

	token, err := helper.GenerateToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	refreshToken, err := helper.GenerateRefreshToken(user, s.cfg)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) HandleLogin(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := s.Login(&req)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return helper.SuccessResponse(c, "Login successful", resp)
}

func (s *AuthService) HandleGetProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	user, err := s.GetProfile(userID)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return helper.SuccessResponse(c, "Profile retrieved", user)
}

func (s *AuthService) HandleRefreshToken(c *fiber.Ctx) error {
	var req model.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := s.RefreshToken(req.RefreshToken)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusUnauthorized, err.Error())
	}

	return helper.SuccessResponse(c, "Token refreshed", resp)
}

func (s *AuthService) HandleLogout(c *fiber.Ctx) error {
	return helper.SuccessResponse(c, "Logout successful", nil)
}
