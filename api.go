package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/api/register", makeHTTPHandlerFunc(s.registerStudentsToTeacher)).Methods("POST")
	router.HandleFunc("/api/commonstudents", makeHTTPHandlerFunc(s.getCommonStudents)).Methods("GET")
	router.HandleFunc("/api/suspend", makeHTTPHandlerFunc(s.suspendStudent)).Methods("POST")
	router.HandleFunc("/api/retrievefornotifications", makeHTTPHandlerFunc(s.studentsToGetNotification)).Methods("POST")

	log.Println("JSON API running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

// Assignment API functions
func (s *APIServer) registerStudentsToTeacher(w http.ResponseWriter, r *http.Request) error {

	registerStudentsToTeacherReq := new(RegisterStudentsToTeacherRequest)
	if err := json.NewDecoder(r.Body).Decode(registerStudentsToTeacherReq); err != nil {
		return err
	}

	// check if the students exist in the database
	exists, err := s.store.TeacherExists(registerStudentsToTeacherReq.TeacherEmail)
	if err != nil {
		return err
	}
	// if entry does not exist, then i create a new entry
	if !exists {
		teacher := NewTeacher(registerStudentsToTeacherReq.TeacherEmail)
		err := s.store.CreateTeacher(teacher)
		if err != nil {
			return err
		}
	}

	for _, studentEmail := range registerStudentsToTeacherReq.StudentEmails {
		// check if the students exist in the database
		exists, err := s.store.StudentExists(studentEmail)
		if err != nil {
			return err
		}
		// if entry does not exist, then i create a new entry
		if !exists {
			student := NewStudent(studentEmail)
			err := s.store.CreateStudent(student)
			if err != nil {
				return err
			}
		}
	}

	for _, studentEmail := range registerStudentsToTeacherReq.StudentEmails {
		// check if the students exist in the database
		exists, err := s.store.TeacherStudentExists(registerStudentsToTeacherReq.TeacherEmail, studentEmail)
		if err != nil {
			return err
		}
		// if entry does not exist, then i create a new entry
		if !exists {
			teacherstudent := NewTeacherStudent(registerStudentsToTeacherReq.TeacherEmail, studentEmail)
			err := s.store.CreateTeacherStudent(teacherstudent)
			if err != nil {
				return err
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *APIServer) getCommonStudents(w http.ResponseWriter, r *http.Request) error {
	// Parse query parameters
	query := r.URL.Query()
	// Get slice of teacher emails
	teachers := query["teacher"]

	commonStudents, err := s.store.GetCommonStudentsOfTeachers(teachers)
	if err != nil {
		return err
	}

	commonStudentsResponse := CommonStudentsResponse{}
	commonStudentsResponse.StudentEmails = commonStudents
	return WriteJSON(w, http.StatusOK, commonStudentsResponse)
}

func (s *APIServer) suspendStudent(w http.ResponseWriter, r *http.Request) error {
	suspendStudentReq := new(SuspendStudentRequest)
	if err := json.NewDecoder(r.Body).Decode(suspendStudentReq); err != nil {
		return err
	}

	//check if student exists
	exists, err := s.store.StudentExists(suspendStudentReq.StudentEmail)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("student does not exist")
	}

	err = s.store.UpdateStudentSuspendedState(suspendStudentReq.StudentEmail, true)

	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *APIServer) studentsToGetNotification(w http.ResponseWriter, r *http.Request) error {
	studentsToGetNotificationReq := new(StudentsToGetNotificationRequest)
	if err := json.NewDecoder(r.Body).Decode(studentsToGetNotificationReq); err != nil {
		return err
	}

	notifiedStudentsResponse := NotifiedStudentsResponse{}
	// get all students under teacher
	students, err := s.store.GetStudentsAssignedToTeacher(studentsToGetNotificationReq.TeacherEmail)
	if err != nil {
		return err
	}
	// check if student is suspended
	for _, student := range students {
		isSuspended, err := s.store.IsStudentSuspended(student)
		if err != nil {
			return err
		}
		if !isSuspended {
			//fmt.Println(students)
			notifiedStudentsResponse.StudentEmails = append(notifiedStudentsResponse.StudentEmails, student)
		}
	}

	// get @mentioned students
	// Define the regular expression pattern to match email addresses
	emailPattern := `@([\w.%+-]+@[\w.-]+\.[a-zA-Z]{2,})` //`\b[\w.%+-]+@[\w.-]+\.[a-zA-Z]{2,}\b`

	// Compile the regular expression pattern
	re := regexp.MustCompile(emailPattern)

	// Find all matches of the pattern in the input string
	mentionedStudents := re.FindAllString(studentsToGetNotificationReq.NotificationString, -1)

	// Print the extracted student emails
	for _, studentEmail := range mentionedStudents {
		// remove the first @ of the mention to get the email
		studentEmail = studentEmail[1:]

		// check if student exist in database
		exists, err := s.store.StudentExists(studentEmail)
		if err != nil {
			return err
		}
		if exists {
			// check if student is suspended
			isSuspended, err := s.store.IsStudentSuspended(studentEmail)
			if err != nil {
				return err
			}
			if !isSuspended {
				// check for duplicates
				hasDuplicate := StringExistsInArray(studentEmail, notifiedStudentsResponse.StudentEmails)
				if !hasDuplicate {
					notifiedStudentsResponse.StudentEmails = append(notifiedStudentsResponse.StudentEmails, studentEmail)
				}
			}
		}
		// else {
		// 	return fmt.Errorf("@mentioned student %s does not exist in database", student_email)
		// }
	}

	//fmt.Println(studentsToGetNotificationReq.TeacherEmail)
	//fmt.Println(studentsToGetNotificationReq.NotificationString)

	return WriteJSON(w, http.StatusOK, notifiedStudentsResponse)
}

// utility functions
// simplifies the process of writing a JSON response to an HTTP request
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

// convert my api Functions to functions of type http.HandlerFunc.
// errors are handled here
func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			// handle error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func StringExistsInArray(input string, input_array []string) bool {
	for _, s := range input_array {
		if s == input {
			return true
		}
	}
	return false
}
