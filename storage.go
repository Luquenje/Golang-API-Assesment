package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateTeacher(*Teacher) error
	GetTeachers() ([]*Teacher, error)
	GetTeacherByEmail(string) (*Teacher, error)
	TeacherExists(string) (bool, error)

	CreateStudent(*Student) error
	UpdateStudentSuspendedState(string, bool) error
	GetStudents() ([]*Student, error)
	GetStudentByEmail(string) (*Student, error)
	StudentExists(string) (bool, error)
	IsStudentSuspended(string) (bool, error)

	CreateTeacherStudent(*TeacherStudent) error
	GetTeacherStudentByEmail(string, string) (*TeacherStudent, error)
	GetStudentsAssignedToTeacher(string) ([]string, error)
	TeacherStudentExists(string, string) (bool, error)

	GetCommonStudentsOfTeachers([]string) ([]string, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr, exists := os.LookupEnv("POSTGRESQL_CONNECTION_STRING")
	if !exists {
		fmt.Println("POSTGRESQL_CONNECTION_STRING env var does not exist")
		connStr = "user=postgres dbname=postgres password=P@ssw0rd sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	err := s.createTeacherTable()
	if err != nil {
		return err
	}
	err = s.createStudentTable()
	if err != nil {
		return err
	}
	err = s.createTeacherStudentTable()
	if err != nil {
		return err
	}

	return nil
}

// Teacher Queries
func (s *PostgresStore) createTeacherTable() error {
	query := `create table if not exists Teacher (
		email VARCHAR(255) PRIMARY KEY,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateTeacher(teacher *Teacher) error {
	query := `
		INSERT INTO Teacher (email, created_at)
		VALUES ($1, $2);`

	_, err := s.db.Query(query, teacher.Email, teacher.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetTeacherByEmail(email string) (*Teacher, error) {
	query := `SELECT * FROM Teacher WHERE email = $1`
	rows, err := s.db.Query(query, email)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		teacher, err := scanIntoTeacher(rows)
		if err != nil {
			return nil, err
		}
		return teacher, nil
	}

	return nil, fmt.Errorf("entry does not exist")
}

func (s *PostgresStore) GetTeachers() ([]*Teacher, error) {
	query := `SELECT * FROM Teacher`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	teachers := []*Teacher{}
	for rows.Next() {
		teacher, err := scanIntoTeacher(rows)

		if err != nil {
			return nil, err
		}

		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

func (s *PostgresStore) TeacherExists(teacheEmail string) (bool, error) {
	query := `SELECT COUNT(1) FROM Teacher WHERE email = $1`

	var count int
	err := s.db.QueryRow(query, teacheEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func scanIntoTeacher(rows *sql.Rows) (*Teacher, error) {
	teacher := new(Teacher)
	err := rows.Scan(
		&teacher.Email,
		&teacher.CreatedAt)

	return teacher, err
}

// Student Queries
func (s *PostgresStore) createStudentTable() error {

	query := `create table if not exists Student (
		email VARCHAR(255) PRIMARY KEY,
		is_suspended bool,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateStudent(student *Student) error {
	query := `
		INSERT INTO Student (email, is_suspended, created_at)
		VALUES ($1, $2, $3);`

	res, err := s.db.Query(query, student.Email, student.IsSuspended, student.CreatedAt)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", res)

	return nil
}

func (s *PostgresStore) GetStudentByEmail(email string) (*Student, error) {
	query := `SELECT * FROM Student WHERE email = $1`
	rows, err := s.db.Query(query, email)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		student, err := scanIntoStudent(rows)
		if err != nil {
			return nil, err
		}
		return student, nil
	}

	return nil, fmt.Errorf("entry does not exist")
}

func (s *PostgresStore) GetStudents() ([]*Student, error) {
	query := `SELECT * FROM Student`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	students := []*Student{}
	for rows.Next() {
		student, err := scanIntoStudent(rows)

		if err != nil {
			return nil, err
		}

		students = append(students, student)
	}

	return students, nil
}

func (s *PostgresStore) UpdateStudentSuspendedState(email string, is_suspended bool) error {
	query := `UPDATE Student SET is_suspended = $1 WHERE email = $2;`
	_, err := s.db.Exec(query, is_suspended, email)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) StudentExists(studentEmail string) (bool, error) {
	query := `SELECT COUNT(1) FROM Student WHERE email = $1`

	var count int
	err := s.db.QueryRow(query, studentEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func (s *PostgresStore) IsStudentSuspended(studentEmail string) (bool, error) {
	query := `SELECT * FROM Student WHERE email = $1`
	rows, err := s.db.Query(query, studentEmail)
	if err != nil {
		return false, err
	}

	for rows.Next() {
		student, err := scanIntoStudent(rows)
		if err != nil {
			return false, err
		}
		return student.IsSuspended, nil
	}
	return false, nil
}

func scanIntoStudent(rows *sql.Rows) (*Student, error) {
	student := new(Student)
	err := rows.Scan(
		&student.Email,
		&student.IsSuspended,
		&student.CreatedAt)

	return student, err
}

// TeacherStudent queries
func (s *PostgresStore) createTeacherStudentTable() error {

	query := `create table if not exists TeacherStudent (
		teacher_email VARCHAR(255),
    	student_email VARCHAR(255),
		created_at timestamp,
    	FOREIGN KEY (teacher_email) REFERENCES Teacher(email),
    	FOREIGN KEY (student_email) REFERENCES Student(email),
    	PRIMARY KEY (teacher_email, student_email)
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateTeacherStudent(teacherstudent *TeacherStudent) error {
	query := `INSERT INTO TeacherStudent (teacher_email, student_email, created_at)
	VALUES ($1, $2, $3);`

	_, err := s.db.Query(query, teacherstudent.TeacherEmail, teacherstudent.StudentEmail, teacherstudent.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetTeacherStudentByEmail(teacherEmail string, studentEmail string) (*TeacherStudent, error) {
	query := `SELECT * FROM TeacherStudent WHERE teacher_email = $1 AND student_email = $2`

	rows, err := s.db.Query(query, teacherEmail, studentEmail)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		teacherstudent, err := scanIntoTeacherStudent(rows)
		if err != nil {
			return nil, err
		}
		return teacherstudent, nil
	}

	return nil, fmt.Errorf("entry does not exist")
}

func (s *PostgresStore) GetStudentsAssignedToTeacher(teacherEmail string) ([]string, error) {
	query := `SELECT * FROM TeacherStudent WHERE teacher_email = $1`
	rows, err := s.db.Query(query, teacherEmail)
	if err != nil {
		return nil, err
	}

	studentEmails := []string{}
	for rows.Next() {
		teacherstudent, err := scanIntoTeacherStudent(rows)
		if err != nil {
			return nil, err
		}
		studentEmails = append(studentEmails, teacherstudent.StudentEmail)
	}

	return studentEmails, nil
}

func (s *PostgresStore) TeacherStudentExists(teacherEmail string, studentEmail string) (bool, error) {
	query := `SELECT COUNT(1) FROM TeacherStudent WHERE teacher_email = $1 AND student_email = $2`

	var count int
	err := s.db.QueryRow(query, teacherEmail, studentEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func scanIntoTeacherStudent(rows *sql.Rows) (*TeacherStudent, error) {
	teacherstudent := new(TeacherStudent)
	err := rows.Scan(
		&teacherstudent.TeacherEmail,
		&teacherstudent.StudentEmail,
		&teacherstudent.CreatedAt)

	return teacherstudent, err
}

// Specific queries
func (s *PostgresStore) GetCommonStudentsOfTeachers(teacherEmails []string) ([]string, error) {
	if len(teacherEmails) == 0 {
		return nil, fmt.Errorf("no teacher emails provided")
	}

	emailsToQueryString := "'" + strings.Join(teacherEmails, "', '") + "'"

	query := fmt.Sprintf(`SELECT Student.email AS common_student
	FROM Student
	JOIN TeacherStudent ts ON Student.email = ts.student_email
	JOIN Teacher ON ts.teacher_email = Teacher.email
	WHERE Teacher.email IN (%s)
	GROUP BY Student.email
	HAVING COUNT(DISTINCT ts.teacher_email) = %d;`, emailsToQueryString, len(teacherEmails))

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	commonStudents := []string{}
	for rows.Next() {
		studentEmail := new(string)
		if err := rows.Scan(&studentEmail); err != nil {
			return nil, err
		}
		commonStudents = append(commonStudents, *studentEmail)
	}

	return commonStudents, nil
}
