package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func TestRegister(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	requestBody := []byte(`{
        "teacher": "teacherken@gmail.com",
        "students": [
            "studentjon@gmail.com",
            "studenthon@gmail.com",
			"student_only_under_teacher_ken@gmail.com"
        ]
    }`)

	requestBody2 := []byte(`{
        "teacher": "teacherjoe@gmail.com",
        "students": [
            "studentjon@gmail.com",
            "studenthon@gmail.com",
			"student_only_under_teacher_joe@gmail.com"
        ]
    }`)

	// Create a new HTTP request with the JSON body
	req, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.registerStudentsToTeacher))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	// Create a new HTTP request with the JSON body
	req, err = http.NewRequest("POST", "/api/register", bytes.NewBuffer(requestBody2))
	if err != nil {
		t.Fatal(err)
		return
	}
	rr = httptest.NewRecorder()

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	fmt.Println("--- Passed Register Test")
}

func TestCommonStudents1(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	// Define the query parameter
	queryParam := "teacher=teacherken%40gmail.com"

	// Create a new HTTP request with the GET method and the query parameter
	req, err := http.NewRequest("GET", "/api/commonstudents?"+queryParam, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.getCommonStudents))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	// Decode the response body
	var responseBody map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Errorf("failed to decode response body: %v", err)
		return
	}

	// Define expected JSON structure or values
	expected := map[string]interface{}{
		"students": []interface{}{
			"student_only_under_teacher_ken@gmail.com",
			"studenthon@gmail.com",
			"studentjon@gmail.com",
		},
	}

	// Compare response body with expected
	if !reflect.DeepEqual(responseBody, expected) {
		t.Errorf("handler returned unexpected response body: got %v, want %v",
			responseBody, expected)
		return
	}

	fmt.Println("--- Passed CommonStudent1 Test")
}

func TestCommonStudents2(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	// Define the query parameter
	queryParam := "teacher=teacherken%40gmail.com&teacher=teacherjoe%40gmail.com"

	// Create a new HTTP request with the GET method and the query parameter
	req, err := http.NewRequest("GET", "/api/commonstudents?"+queryParam, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.getCommonStudents))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	// Decode the response body
	var responseBody CommonStudentsResponse
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Errorf("failed to decode response body: %v", err)
		return
	}

	// Define expected JSON structure or values
	expected := CommonStudentsResponse{StudentEmails: []string{
		"studenthon@gmail.com",
		"studentjon@gmail.com"}}

	// Compare response body with expected
	if !reflect.DeepEqual(responseBody, expected) {
		t.Errorf("handler returned unexpected response body: got %v, want %v",
			responseBody, expected)
		return
	}

	fmt.Println("--- Passed CommonStudent2 Test")
}

func TestSuspend(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	requestBody := []byte(`{
        "student": "studentjon@gmail.com"
    }`)

	// Create a new HTTP request with the JSON body
	req, err := http.NewRequest("POST", "/api/suspend", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.suspendStudent))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	fmt.Println("--- Passed Suspend Test")
}

func TestNotification1(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	requestBody := []byte(`{
		"teacher":  "teacherken@gmail.com",
		"notification": "Hello students! @student_only_under_teacher_joe@gmail.com"
	}`)

	// Create a new HTTP request with the JSON body
	req, err := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.studentsToGetNotification))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	// Decode the response body
	var responseBody NotifiedStudentsResponse
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Errorf("failed to decode response body: %v", err)
		return
	}

	// Define expected JSON structure or values
	expected := NotifiedStudentsResponse{StudentEmails: []string{
		"student_only_under_teacher_ken@gmail.com",
		"studenthon@gmail.com",
		"student_only_under_teacher_joe@gmail.com"}}

	// Compare response body with expected
	if !reflect.DeepEqual(responseBody, expected) {
		t.Errorf("handler returned unexpected response body: got %v, want %v",
			responseBody, expected)
		return
	}

	fmt.Println("--- Passed Notification1 Test")
}

func TestNotification2(t *testing.T) {
	store, err := NewPostgresStore()
	if err != nil {
		t.Fatal(err)
		return
	}

	if err := store.Init(); err != nil {
		t.Fatal(err)
		return
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "3000"
	}
	server := NewAPIServer(":"+port, store)

	requestBody := []byte(`{
		"teacher":  "teacherken@gmail.com",
		"notification": "Hey everybody"
	}`)

	// Create a new HTTP request with the JSON body
	req, err := http.NewRequest("POST", "/api/retrievefornotifications", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
		return
	}

	// Create a recorder to record the response
	rr := httptest.NewRecorder()

	// Define a handler function for testing
	handler := http.HandlerFunc(makeHTTPHandlerFunc(server.studentsToGetNotification))

	// Serve the HTTP request to the handler
	handler.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
		return
	}

	// Decode the response body
	var responseBody NotifiedStudentsResponse
	if err := json.NewDecoder(rr.Body).Decode(&responseBody); err != nil {
		t.Errorf("failed to decode response body: %v", err)
		return
	}

	// Define expected JSON structure or values
	expected := NotifiedStudentsResponse{StudentEmails: []string{
		"student_only_under_teacher_ken@gmail.com",
		"studenthon@gmail.com"}}

	// Compare response body with expected
	if !reflect.DeepEqual(responseBody, expected) {
		t.Errorf("handler returned unexpected response body: got %v, want %v",
			responseBody, expected)
		return
	}

	fmt.Println("--- Passed Notification2 Test")
}
