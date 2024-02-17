package main

import (
	"time"
)

type ApiError struct {
	Error string
}

// request types
type RegisterStudentsToTeacherRequest struct {
	TeacherEmail  string   `json:"teacher"`
	StudentEmails []string `json:"students"`
}

type CreateTeacherRequest struct {
	Email string `json:"email"`
}

type SuspendStudentRequest struct {
	StudentEmail string `json:"student"`
}

type StudentsToGetNotificationRequest struct {
	TeacherEmail       string `json:"teacher"`
	NotificationString string `json:"notification"`
}

// response types
type CommonStudentsResponse struct {
	StudentEmails []string `json:"students"`
}
type NotifiedStudentsResponse struct {
	StudentEmails []string `json:"recipients"`
}

// teacher
type Teacher struct {
	// ID        int       `json:id`
	Email     string    `json:email`
	CreatedAt time.Time `json:created_at`
}

func NewTeacher(email string) *Teacher {
	return &Teacher{
		Email:     email,
		CreatedAt: time.Now().UTC(),
	}
}

// student
type Student struct {
	// ID        int       `json:id`
	Email       string    `json:email`
	IsSuspended bool      `json:is_suspended`
	CreatedAt   time.Time `json:created_at`
}

func NewStudent(email string) *Student {
	return &Student{
		Email:       email,
		IsSuspended: false,
		CreatedAt:   time.Now().UTC(),
	}
}

// teacher-student
type TeacherStudent struct {
	// ID        int       `json:id`
	TeacherEmail string    `json:teacher_email`
	StudentEmail string    `json:student_email`
	CreatedAt    time.Time `json:created_at`
}

func NewTeacherStudent(teacher_email, student_email string) *TeacherStudent {
	return &TeacherStudent{
		TeacherEmail: teacher_email,
		StudentEmail: student_email,
		CreatedAt:    time.Now().UTC(),
	}
}
