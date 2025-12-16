package service

import (
	"errors"
	"projek_uas/model"
	"projek_uas/repository"
)

type AchievementService struct {
	achievementRepo *repository.AchievementRepository
	studentRepo     *repository.StudentRepository
}

func NewAchievementService(achievementRepo *repository.AchievementRepository, studentRepo *repository.StudentRepository) *AchievementService {
	return &AchievementService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
	}
}

func (s *AchievementService) CreateAchievement(userID string, req *model.CreateAchievementRequest) (*model.AchievementReference, error) {
	// Get student
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.New("student profile not found")
	}

	// Create achievement in MongoDB
	achievement := &model.Achievement{
		StudentID:       student.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		Points:          req.Points,
		Attachments:     []model.Attachment{},
	}

	if err := s.achievementRepo.CreateMongo(achievement); err != nil {
		return nil, err
	}

	// Create reference in PostgreSQL
	ref := &model.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             "draft",
	}

	if err := s.achievementRepo.CreateReference(ref); err != nil {
		return nil, err
	}

	ref.Achievement = achievement
	return ref, nil
}

func (s *AchievementService) GetAchievements(userID, roleName string, status string, page, limit int) ([]*model.AchievementReference, *model.Pagination, error) {
	offset := (page - 1) * limit
	var studentIDs []string
	var err error

	if roleName == "Mahasiswa" {
		student, err := s.studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, nil, err
		}
		if student != nil {
			studentIDs = []string{student.ID}
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := repository.NewLecturerRepository().FindByUserID(userID)
		if err != nil {
			return nil, nil, err
		}
		if lecturer != nil {
			studentIDs, err = s.studentRepo.GetStudentsByAdvisorID(lecturer.ID)
			if err != nil {
				return nil, nil, err
			}
		}
	}
	// Admin gets all (empty studentIDs)

	refs, total, err := s.achievementRepo.GetReferences(studentIDs, status, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	// Fetch achievement details from MongoDB
	for _, ref := range refs {
		achievement, err := s.achievementRepo.FindMongoByID(ref.MongoAchievementID)
		if err == nil {
			ref.Achievement = achievement
		}
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

	return refs, pagination, nil
}

func (s *AchievementService) GetAchievementByID(id, userID, roleName string) (*model.AchievementReference, error) {
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, errors.New("achievement not found")
	}

	// Check permissions
	if roleName == "Mahasiswa" {
		student, err := s.studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if student == nil || student.ID != ref.StudentID {
			return nil, errors.New("unauthorized")
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := repository.NewLecturerRepository().FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lecturer != nil {
			studentIDs, err := s.studentRepo.GetStudentsByAdvisorID(lecturer.ID)
			if err != nil {
				return nil, err
			}
			authorized := false
			for _, sid := range studentIDs {
				if sid == ref.StudentID {
					authorized = true
					break
				}
			}
			if !authorized {
				return nil, errors.New("unauthorized")
			}
		}
	}

	// Fetch achievement details
	achievement, err := s.achievementRepo.FindMongoByID(ref.MongoAchievementID)
	if err != nil {
		return nil, err
	}
	ref.Achievement = achievement

	return ref, nil
}

func (s *AchievementService) UpdateAchievement(id, userID string, req *model.UpdateAchievementRequest) error {
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	// Check if user owns this achievement
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	// Only allow update if status is draft or rejected
	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("cannot update achievement in current status")
	}

	// Update MongoDB
	achievement := &model.Achievement{
		Title:       req.Title,
		Description: req.Description,
		Details:     req.Details,
		Tags:        req.Tags,
		Points:      req.Points,
	}

	return s.achievementRepo.UpdateMongo(ref.MongoAchievementID, achievement)
}

func (s *AchievementService) DeleteAchievement(id, userID string) error {
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	// Check if user owns this achievement
	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	// Only allow delete if status is draft
	if ref.Status != "draft" {
		return errors.New("can only delete draft achievements")
	}

	// Delete from MongoDB
	if err := s.achievementRepo.DeleteMongo(ref.MongoAchievementID); err != nil {
		return err
	}

	// Delete reference
	return s.achievementRepo.DeleteReference(id)
}

func (s *AchievementService) SubmitForVerification(id, userID string) error {
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("achievement already submitted")
	}

	return s.achievementRepo.UpdateReferenceStatus(id, "submitted", nil, nil)
}

func (s *AchievementService) VerifyAchievement(id, userID string, req *model.VerifyAchievementRequest) error {
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	if ref.Status != "submitted" {
		return errors.New("achievement is not in submitted status")
	}

	if req.Action == "verify" {
		return s.achievementRepo.UpdateReferenceStatus(id, "verified", &userID, nil)
	} else if req.Action == "reject" {
		return s.achievementRepo.UpdateReferenceStatus(id, "rejected", &userID, &req.Note)
	}

	return errors.New("invalid action")
}

func (s *AchievementService) GetStatistics(userID, roleName string) (map[string]interface{}, error) {
	var studentIDs []string

	if roleName == "Mahasiswa" {
		student, err := s.studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if student != nil {
			studentIDs = []string{student.ID}
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := repository.NewLecturerRepository().FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lecturer != nil {
			var err error
			studentIDs, err = s.studentRepo.GetStudentsByAdvisorID(lecturer.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	return s.achievementRepo.GetStatistics(studentIDs)
}
