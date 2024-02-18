package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	db     *sql.DB
	dbPool *pgxpool.Pool
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr, exists := os.LookupEnv("POSTGRESQL_CONNECTION_STRING")
	if !exists {
		fmt.Println("POSTGRESQL_CONNECTION_STRING env var does not exist")
		connStr = "user=postgres dbname=postgres password=P@ssw0rd sslmode=disable"
	}
	//db, err := sql.Open("postgres", connStr)

	// if err != nil {
	// 	return nil, err
	// }
	// if err := db.Ping(); err != nil {
	// 	return nil, err
	// }

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of connections in the pool
	poolConfig.MaxConns = 5

	// Create the database connection pool
	dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &PostgresStore{
		dbPool: dbPool,
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `create table if not exists Teacher (
		email VARCHAR(255) PRIMARY KEY,
		created_at timestamp
	)`

	_, err = conn.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) CreateTeacher(teacher *Teacher) error {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `
		INSERT INTO Teacher (email, created_at)
		VALUES ($1, $2);`

	_, err = conn.Query(context.Background(), query, teacher.Email, teacher.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetTeacherByEmail(email string) (*Teacher, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM Teacher WHERE email = $1`
	rows, err := conn.Query(context.Background(), query, email)
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM Teacher`
	rows, err := conn.Query(context.Background(), query)
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Release()

	query := `SELECT COUNT(1) FROM Teacher WHERE email = $1`

	var count int
	err = conn.QueryRow(context.Background(), query, teacheEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func scanIntoTeacher(rows pgx.Rows) (*Teacher, error) {
	teacher := new(Teacher)
	err := rows.Scan(
		&teacher.Email,
		&teacher.CreatedAt)

	return teacher, err
}

// Student Queries
func (s *PostgresStore) createStudentTable() error {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `create table if not exists Student (
		email VARCHAR(255) PRIMARY KEY,
		is_suspended bool,
		created_at timestamp
	)`

	_, err = conn.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) CreateStudent(student *Student) error {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()
	query := `
		INSERT INTO Student (email, is_suspended, created_at)
		VALUES ($1, $2, $3);`

	_, err = conn.Query(context.Background(), query, student.Email, student.IsSuspended, student.CreatedAt)

	if err != nil {
		return err
	}

	//fmt.Printf("%+v\n", res)

	return nil
}

func (s *PostgresStore) GetStudentByEmail(email string) (*Student, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM Student WHERE email = $1`
	rows, err := conn.Query(context.Background(), query, email)
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM Student`
	rows, err := conn.Query(context.Background(), query)
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `UPDATE Student SET is_suspended = $1 WHERE email = $2;`
	_, err = conn.Exec(context.Background(), query, is_suspended, email)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) StudentExists(studentEmail string) (bool, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Release()

	query := `SELECT COUNT(1) FROM Student WHERE email = $1`

	var count int
	err = conn.QueryRow(context.Background(), query, studentEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func (s *PostgresStore) IsStudentSuspended(studentEmail string) (bool, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Release()

	query := `SELECT * FROM Student WHERE email = $1`
	rows, err := conn.Query(context.Background(), query, studentEmail)
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

func scanIntoStudent(rows pgx.Rows) (*Student, error) {
	student := new(Student)
	err := rows.Scan(
		&student.Email,
		&student.IsSuspended,
		&student.CreatedAt)

	return student, err
}

// TeacherStudent queries
func (s *PostgresStore) createTeacherStudentTable() error {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `create table if not exists TeacherStudent (
		teacher_email VARCHAR(255),
    	student_email VARCHAR(255),
		created_at timestamp,
    	FOREIGN KEY (teacher_email) REFERENCES Teacher(email),
    	FOREIGN KEY (student_email) REFERENCES Student(email),
    	PRIMARY KEY (teacher_email, student_email)
	)`

	_, err = conn.Exec(context.Background(), query)
	return err
}

func (s *PostgresStore) CreateTeacherStudent(teacherstudent *TeacherStudent) error {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	query := `INSERT INTO TeacherStudent (teacher_email, student_email, created_at)
	VALUES ($1, $2, $3);`

	_, err = conn.Query(context.Background(), query, teacherstudent.TeacherEmail, teacherstudent.StudentEmail, teacherstudent.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) GetTeacherStudentByEmail(teacherEmail string, studentEmail string) (*TeacherStudent, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM TeacherStudent WHERE teacher_email = $1 AND student_email = $2`

	rows, err := conn.Query(context.Background(), query, teacherEmail, studentEmail)

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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := `SELECT * FROM TeacherStudent WHERE teacher_email = $1`
	rows, err := conn.Query(context.Background(), query, teacherEmail)
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
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Release()

	query := `SELECT COUNT(1) FROM TeacherStudent WHERE teacher_email = $1 AND student_email = $2`

	var count int
	err = conn.QueryRow(context.Background(), query, teacherEmail, studentEmail).Scan(&count) // queries the count of rows selected
	if err != nil {
		return false, err
	}

	exists := count > 0
	return exists, nil
}

func scanIntoTeacherStudent(rows pgx.Rows) (*TeacherStudent, error) {
	teacherstudent := new(TeacherStudent)
	err := rows.Scan(
		&teacherstudent.TeacherEmail,
		&teacherstudent.StudentEmail,
		&teacherstudent.CreatedAt)

	return teacherstudent, err
}

// Specific queries
func (s *PostgresStore) GetCommonStudentsOfTeachers(teacherEmails []string) ([]string, error) {
	conn, err := s.dbPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

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

	rows, err := conn.Query(context.Background(), query)
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
