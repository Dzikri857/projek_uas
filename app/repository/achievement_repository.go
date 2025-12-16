package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"projek_uas/database"
	"projek_uas/helper"
	"projek_uas/app/model"
)

type AchievementRepository struct{}

func NewAchievementRepository() *AchievementRepository {
	return &AchievementRepository{}
}

// MongoDB operations
func (r *AchievementRepository) CreateMongo(achievement *model.Achievement) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	result, err := database.MongoDB.Collection("achievements").InsertOne(ctx, achievement)
	if err != nil {
		return err
	}

	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AchievementRepository) FindMongoByID(id string) (*model.Achievement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var achievement model.Achievement
	err = database.MongoDB.Collection("achievements").FindOne(ctx, bson.M{"_id": objectID}).Decode(&achievement)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

func (r *AchievementRepository) UpdateMongo(id string, achievement *model.Achievement) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	achievement.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"title":           achievement.Title,
			"description":     achievement.Description,
			"details":         achievement.Details,
			"tags":            achievement.Tags,
			"points":          achievement.Points,
			"updatedAt":       achievement.UpdatedAt,
		},
	}

	_, err = database.MongoDB.Collection("achievements").UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *AchievementRepository) DeleteMongo(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = database.MongoDB.Collection("achievements").DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// PostgreSQL operations for achievement references
func (r *AchievementRepository) CreateReference(ref *model.AchievementReference) error {
	query := `
		INSERT INTO achievement_references (student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return database.PostgresDB.QueryRow(query, ref.StudentID, ref.MongoAchievementID, ref.Status).
		Scan(&ref.ID, &ref.CreatedAt, &ref.UpdatedAt)
}

func (r *AchievementRepository) FindReferenceByID(id string) (*model.AchievementReference, error) {
	ref := &model.AchievementReference{}
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at,
		       verified_by, rejection_note, created_at, updated_at
		FROM achievement_references WHERE id = $1
	`
	err := database.PostgresDB.QueryRow(query, id).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ref, err
}

func (r *AchievementRepository) GetReferences(studentIDs []string, status string, limit, offset int) ([]*model.AchievementReference, int64, error) {
	var query string
	var args []interface{}
	argIndex := 1

	if len(studentIDs) > 0 {
		query = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at,
			       verified_by, rejection_note, created_at, updated_at
			FROM achievement_references
			WHERE student_id = ANY($1)
		`
		args = append(args, studentIDs)
		argIndex++
	} else {
		query = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at,
			       verified_by, rejection_note, created_at, updated_at
			FROM achievement_references
			WHERE 1=1
		`
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := database.PostgresDB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var refs []*model.AchievementReference
	for rows.Next() {
		ref := &model.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		refs = append(refs, ref)
	}

	// Count total
	var countQuery string
	var countArgs []interface{}
	if len(studentIDs) > 0 {
		countQuery = "SELECT COUNT(*) FROM achievement_references WHERE student_id = ANY($1)"
		countArgs = append(countArgs, studentIDs)
		if status != "" {
			countQuery += " AND status = $2"
			countArgs = append(countArgs, status)
		}
	} else {
		countQuery = "SELECT COUNT(*) FROM achievement_references WHERE 1=1"
		if status != "" {
			countQuery += " AND status = $1"
			countArgs = append(countArgs, status)
		}
	}

	var total int64
	err = database.PostgresDB.QueryRow(countQuery, countArgs...).Scan(&total)

	return refs, total, err
}

func (r *AchievementRepository) UpdateReferenceStatus(id, status string, verifiedBy *string, note *string) error {
	now := time.Now()
	query := `
		UPDATE achievement_references
		SET status = $1, updated_at = $2
	`
	args := []interface{}{status, now}
	argIndex := 3

	if status == "submitted" {
		query += fmt.Sprintf(", submitted_at = $%d", argIndex)
		args = append(args, now)
		argIndex++
	} else if status == "verified" {
		query += fmt.Sprintf(", verified_at = $%d, verified_by = $%d", argIndex, argIndex+1)
		args = append(args, now, verifiedBy)
		argIndex += 2
	} else if status == "rejected" {
		query += fmt.Sprintf(", verified_by = $%d, rejection_note = $%d", argIndex, argIndex+1)
		args = append(args, verifiedBy, note)
		argIndex += 2
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := database.PostgresDB.Exec(query, args...)
	return err
}

func (r *AchievementRepository) DeleteReference(id string) error {
	_, err := database.PostgresDB.Exec("DELETE FROM achievement_references WHERE id = $1", id)
	return err
}

func (r *AchievementRepository) GetStatistics(studentIDs []string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build match stage
	matchStage := bson.M{}
	if len(studentIDs) > 0 {
		matchStage["studentId"] = bson.M{"$in": studentIDs}
	}

	pipeline := []bson.M{}
	if len(matchStage) > 0 {
		pipeline = append(pipeline, bson.M{"$match": matchStage})
	}

	// Count by type
	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":   "$achievementType",
			"count": bson.M{"$sum": 1},
			"totalPoints": bson.M{"$sum": "$points"},
		},
	})

	cursor, err := database.MongoDB.Collection("achievements").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"by_type": results,
	}

	return stats, nil
}

// HTTP handler methods integrated into repository
func (r *AchievementRepository) HandleCreate(userID string, req *model.CreateAchievementRequest, studentRepo *StudentRepository) (*model.AchievementReference, error) {
	student, err := studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.New("student profile not found")
	}

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

	if err := r.CreateMongo(achievement); err != nil {
		return nil, err
	}

	ref := &model.AchievementReference{
		StudentID:          student.ID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             "draft",
	}

	if err := r.CreateReference(ref); err != nil {
		return nil, err
	}

	ref.Achievement = achievement
	return ref, nil
}

func (r *AchievementRepository) HandleGetAll(userID, roleName, status string, page, limit int, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) ([]*model.AchievementReference, *model.Pagination, error) {
	offset := (page - 1) * limit
	var studentIDs []string
	var err error

	if roleName == "Mahasiswa" {
		student, err := studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, nil, err
		}
		if student != nil {
			studentIDs = []string{student.ID}
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := lecturerRepo.FindByUserID(userID)
		if err != nil {
			return nil, nil, err
		}
		if lecturer != nil {
			studentIDs, err = studentRepo.GetStudentsByAdvisorID(lecturer.ID)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	refs, total, err := r.GetReferences(studentIDs, status, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	for _, ref := range refs {
		achievement, err := r.FindMongoByID(ref.MongoAchievementID)
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

func (r *AchievementRepository) HandleGetByID(id, userID, roleName string, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) (*model.AchievementReference, error) {
	ref, err := r.FindReferenceByID(id)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, errors.New("achievement not found")
	}

	if roleName == "Mahasiswa" {
		student, err := studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if student == nil || student.ID != ref.StudentID {
			return nil, errors.New("unauthorized")
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := lecturerRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lecturer != nil {
			studentIDs, err := studentRepo.GetStudentsByAdvisorID(lecturer.ID)
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

	achievement, err := r.FindMongoByID(ref.MongoAchievementID)
	if err != nil {
		return nil, err
	}
	ref.Achievement = achievement

	return ref, nil
}

func (r *AchievementRepository) HandleUpdate(id, userID string, req *model.UpdateAchievementRequest, studentRepo *StudentRepository) error {
	ref, err := r.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	student, err := studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("cannot update achievement in current status")
	}

	achievement := &model.Achievement{
		Title:       req.Title,
		Description: req.Description,
		Details:     req.Details,
		Tags:        req.Tags,
		Points:      req.Points,
	}

	return r.UpdateMongo(ref.MongoAchievementID, achievement)
}

func (r *AchievementRepository) HandleDelete(id, userID string, studentRepo *StudentRepository) error {
	ref, err := r.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	student, err := studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	if ref.Status != "draft" {
		return errors.New("can only delete draft achievements")
	}

	if err := r.DeleteMongo(ref.MongoAchievementID); err != nil {
		return err
	}

	return r.DeleteReference(id)
}

func (r *AchievementRepository) HandleSubmit(id, userID string, studentRepo *StudentRepository) error {
	ref, err := r.FindReferenceByID(id)
	if err != nil {
		return err
	}
	if ref == nil {
		return errors.New("achievement not found")
	}

	student, err := studentRepo.FindByUserID(userID)
	if err != nil {
		return err
	}
	if student == nil || student.ID != ref.StudentID {
		return errors.New("unauthorized")
	}

	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("achievement already submitted")
	}

	return r.UpdateReferenceStatus(id, "submitted", nil, nil)
}

func (r *AchievementRepository) HandleVerify(id, userID string, req *model.VerifyAchievementRequest) error {
	ref, err := r.FindReferenceByID(id)
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
		return r.UpdateReferenceStatus(id, "verified", &userID, nil)
	} else if req.Action == "reject" {
		return r.UpdateReferenceStatus(id, "rejected", &userID, &req.Note)
	}

	return errors.New("invalid action")
}

func (r *AchievementRepository) HandleStatistics(userID, roleName string, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) (map[string]interface{}, error) {
	var studentIDs []string

	if roleName == "Mahasiswa" {
		student, err := studentRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if student != nil {
			studentIDs = []string{student.ID}
		}
	} else if roleName == "Dosen Wali" {
		lecturer, err := lecturerRepo.FindByUserID(userID)
		if err != nil {
			return nil, err
		}
		if lecturer != nil {
			var err error
			studentIDs, err = studentRepo.GetStudentsByAdvisorID(lecturer.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	return r.GetStatistics(studentIDs)
}

func (r *AchievementRepository) HandleCreateHTTP(c *fiber.Ctx, studentRepo *StudentRepository) error {
	userID := c.Locals("userID").(string)

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	achievement, err := r.HandleCreate(userID, &req, studentRepo)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Achievement created successfully", achievement)
}

func (r *AchievementRepository) HandleGetAllHTTP(c *fiber.Ctx, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) error {
	userID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)
	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	achievements, pagination, err := r.HandleGetAll(userID, roleName, status, page, limit, studentRepo, lecturerRepo)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.PaginatedResponse(c, achievements, *pagination)
}

func (r *AchievementRepository) HandleGetByIDHTTP(c *fiber.Ctx, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	achievement, err := r.HandleGetByID(id, userID, roleName, studentRepo, lecturerRepo)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusNotFound, err.Error())
	}

	return helper.SuccessResponse(c, "Achievement retrieved", achievement)
}

func (r *AchievementRepository) HandleUpdateHTTP(c *fiber.Ctx, studentRepo *StudentRepository) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := r.HandleUpdate(id, userID, &req, studentRepo); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Achievement updated successfully", nil)
}

func (r *AchievementRepository) HandleDeleteHTTP(c *fiber.Ctx, studentRepo *StudentRepository) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	if err := r.HandleDelete(id, userID, studentRepo); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Achievement deleted successfully", nil)
}

func (r *AchievementRepository) HandleSubmitHTTP(c *fiber.Ctx, studentRepo *StudentRepository) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	if err := r.HandleSubmit(id, userID, studentRepo); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.SuccessResponse(c, "Achievement submitted for verification", nil)
}

func (r *AchievementRepository) HandleVerifyHTTP(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("userID").(string)

	var req model.VerifyAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := r.HandleVerify(id, userID, &req); err != nil {
		return helper.ErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	message := "Achievement verified successfully"
	if req.Action == "reject" {
		message = "Achievement rejected"
	}

	return helper.SuccessResponse(c, message, nil)
}

func (r *AchievementRepository) HandleStatisticsHTTP(c *fiber.Ctx, studentRepo *StudentRepository, lecturerRepo *LecturerRepository) error {
	userID := c.Locals("userID").(string)
	roleName := c.Locals("roleName").(string)

	stats, err := r.HandleStatistics(userID, roleName, studentRepo, lecturerRepo)
	if err != nil {
		return helper.ErrorResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	return helper.SuccessResponse(c, "Statistics retrieved", stats)
}
