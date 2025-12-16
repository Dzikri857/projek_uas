package repository

import (
	"database/sql"
	"projek_uas/database"
	"projek_uas/model"
)

type LecturerRepository struct{}

func NewLecturerRepository() *LecturerRepository {
	return &LecturerRepository{}
}

func (r *LecturerRepository) Create(lecturer *model.Lecturer) error {
	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return database.PostgresDB.QueryRow(
		query,
		lecturer.UserID, lecturer.LecturerID, lecturer.Department,
	).Scan(&lecturer.ID, &lecturer.CreatedAt)
}

func (r *LecturerRepository) FindByUserID(userID string) (*model.Lecturer, error) {
	lecturer := &model.Lecturer{}
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers WHERE user_id = $1
	`
	err := database.PostgresDB.QueryRow(query, userID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return lecturer, err
}
