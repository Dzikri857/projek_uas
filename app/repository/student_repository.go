package repository

import (
	"database/sql"
	"projek_uas/app/model" // Fixed import path to use app/model
	"projek_uas/database"
)

type StudentRepository struct{}

func NewStudentRepository() *StudentRepository {
	return &StudentRepository{}
}

func (r *StudentRepository) Create(student *model.Student) error {
	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return database.PostgresDB.QueryRow(
		query,
		student.UserID, student.StudentID, student.ProgramStudy, student.AcademicYear, student.AdvisorID,
	).Scan(&student.ID, &student.CreatedAt)
}

func (r *StudentRepository) FindByUserID(userID string) (*model.Student, error) {
	student := &model.Student{}
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE user_id = $1
	`
	err := database.PostgresDB.QueryRow(query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy,
		&student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return student, err
}

func (r *StudentRepository) FindByID(id string) (*model.Student, error) {
	student := &model.Student{}
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE id = $1
	`
	err := database.PostgresDB.QueryRow(query, id).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy,
		&student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return student, err
}

func (r *StudentRepository) GetStudentsByAdvisorID(advisorID string) ([]string, error) {
	query := "SELECT id FROM students WHERE advisor_id = $1"
	rows, err := database.PostgresDB.Query(query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id)
	}
	return studentIDs, nil
}

func (r *StudentRepository) UpdateAdvisor(studentID, advisorID string) error {
	query := "UPDATE students SET advisor_id = $1 WHERE id = $2"
	_, err := database.PostgresDB.Exec(query, advisorID, studentID)
	return err
}
