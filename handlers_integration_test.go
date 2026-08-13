package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/models"
	"task-manager-api/repository"
	"task-manager-api/service"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION") == "1" {
		if err := godotenv.Load(); err != nil {
			fmt.Println("failed to load .env file:", err)
			os.Exit(1)
		}
		config = loadConfig()
		gin.SetMode(gin.TestMode)

		db, err := connectToTestDB()
		if err != nil {
			fmt.Println("failed to connect to test database:", err)
			os.Exit(1)
		}
		testDB = db
		if err := truncateTestTables(testDB); err != nil {
			fmt.Println("failed to truncate test tables:", err)
			os.Exit(1)
		}
	}
	code := m.Run()
	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

func connectToTestDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=taskmanager_test sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword,
	)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func truncateTestTables(db *sql.DB) error {
	_, err := db.Exec("TRUNCATE tasks, users RESTART IDENTITY CASCADE")
	return err
}

func requireTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("integration tests skipped: set RUN_INTEGRATION=1 to run against taskmanager_test")
	}
	return testDB
}

func createTestUser(t *testing.T, db *sql.DB, email string) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		"INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id",
		email, "hashed-password",
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	return id
}

func newTestHandler(db *sql.DB) *Handler {
	taskRepo := &repository.TaskRepository{DB: db}
	userRepo := &repository.UserRepo{DB: db}
	return &Handler{
		TaskService: service.Taskservice{TaskRepo: taskRepo},
		UserService: service.UserService{UserRepo: userRepo},
	}
}

func newTestRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.GET("/tasks", AuthMiddleware, h.handleGetTask)
	r.GET("/tasks/:id", AuthMiddleware, h.handleGetTask)
	r.POST("/tasks", AuthMiddleware, h.handlePostTask)
	r.POST("/register", h.handleRegister)
	r.POST("/login", h.handleLogin)
	r.PUT("/tasks/:id", AuthMiddleware, h.handlePutTask)
	r.DELETE("/tasks/:id", AuthMiddleware, h.handleDeleteTask)
	return r
}

func postJSON(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return doRequest(t, r, http.MethodPost, path, body, "")
}

func doRequest(t *testing.T, r *gin.Engine, method, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func testToken(t *testing.T, userID int) string {
	t.Helper()
	token, err := generateToken(userID, "user", config.JWT_SECRET)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func insertTestTask(t *testing.T, db *sql.DB, title string, userID int) int {
	t.Helper()
	var id int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		title, userID,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}
	return id
}

func TestRegisterHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	email := "handler-register@example.com"
	password := "secret123"
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)

	rec := postJSON(t, router, "/register", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if resp["message"] != "User Created" {
		t.Errorf("expected message %q, got %q", "User Created", resp["message"])
	}

	var storedEmail, storedPassword string
	err := db.QueryRow(
		"SELECT email, password FROM users WHERE email = $1", email,
	).Scan(&storedEmail, &storedPassword)
	if err != nil {
		t.Fatalf("failed to query registered user: %v", err)
	}
	if storedEmail != email {
		t.Errorf("expected stored email %q, got %q", email, storedEmail)
	}
	if storedPassword == password {
		t.Errorf("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
		t.Errorf("stored password is not a bcrypt hash of the submitted password: %v", err)
	}
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	email := "handler-duplicate@example.com"
	password := "secret123"
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)

	rec := postJSON(t, router, "/register", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected first registration status 201, got %d; body=%s", rec.Code, rec.Body.String())
	}

	rec = postJSON(t, router, "/register", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate registration status 409, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if resp["error"] != "Email already exists" {
		t.Errorf("expected error %q, got %q", "Email already exists", resp["error"])
	}
}

func TestRegisterHandler_ShortPassword(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	email := "handler-short@example.com"
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, "123")

	rec := postJSON(t, router, "/register", body)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "password must be at least 6 characters") {
		t.Errorf("expected password validation message, got body=%s", rec.Body.String())
	}
}

func TestGetTasksHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-get-tasks@example.com")
	taskID1 := insertTestTask(t, db, "Alpha", userID)
	taskID2 := insertTestTask(t, db, "Beta", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodGet, "/tasks", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var tasks []models.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d: %v", len(tasks), tasks)
	}

	ids := make(map[int]bool, len(tasks))
	titles := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		ids[task.ID] = true
		titles[task.Title] = true
	}
	if !ids[taskID1] || !ids[taskID2] {
		t.Errorf("expected task IDs %d and %d, got %v", taskID1, taskID2, ids)
	}
	if !titles["Alpha"] || !titles["Beta"] {
		t.Errorf("expected tasks Alpha and Beta, got %v", titles)
	}
}

func TestGetTasksHandler_NoToken(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	rec := doRequest(t, router, http.MethodGet, "/tasks", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Errorf("expected unauthorized message, got body=%s", rec.Body.String())
	}
}

func TestGetTasksHandler_UserIsolation(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userA := createTestUser(t, db, "ht-isolate-a@example.com")
	userB := createTestUser(t, db, "ht-isolate-b@example.com")

	insertTestTask(t, db, "A task one", userA)
	insertTestTask(t, db, "A task two", userA)
	insertTestTask(t, db, "B task one", userB)
	insertTestTask(t, db, "B task two", userB)

	token := testToken(t, userA)
	rec := doRequest(t, router, http.MethodGet, "/tasks", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var tasks []models.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected only user A's 2 tasks, got %d: %v", len(tasks), tasks)
	}
	for _, task := range tasks {
		if task.Title == "B task one" || task.Title == "B task two" {
			t.Errorf("user B's task leaked into user A's response: %v", task)
		}
	}
}

func TestGetTaskByIDHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-getbyid@example.com")
	taskID := insertTestTask(t, db, "Single task", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodGet, fmt.Sprintf("/tasks/%d", taskID), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var task models.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if task.ID != taskID {
		t.Errorf("expected ID %d, got %d", taskID, task.ID)
	}
	if task.Title != "Single task" {
		t.Errorf("expected title %q, got %q", "Single task", task.Title)
	}
}

func TestGetTaskByIDHandler_NoToken(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	rec := doRequest(t, router, http.MethodGet, "/tasks/1", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Errorf("expected unauthorized message, got body=%s", rec.Body.String())
	}
}

func TestGetTaskByIDHandler_WrongOwner(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	owner := createTestUser(t, db, "ht-getbyid-owner@example.com")
	intruder := createTestUser(t, db, "ht-getbyid-intruder@example.com")
	taskID := insertTestTask(t, db, "Owned task", owner)
	token := testToken(t, intruder)

	rec := doRequest(t, router, http.MethodGet, fmt.Sprintf("/tasks/%d", taskID), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Task not found") {
		t.Errorf("expected task-not-found message, got body=%s", rec.Body.String())
	}
}

func TestGetTaskByIDHandler_NotFound(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-getbyid-nf@example.com")
	taskID := insertTestTask(t, db, "Existing task", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodGet, fmt.Sprintf("/tasks/%d", taskID+100000), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostTaskHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-post@example.com")
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodPost, "/tasks", `{"title":"Created task"}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var task models.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if task.ID == 0 {
		t.Errorf("expected non-zero task ID")
	}
	if task.Title != "Created task" {
		t.Errorf("expected title %q, got %q", "Created task", task.Title)
	}

	var storedTitle string
	var storedUserID int
	err := db.QueryRow(
		"SELECT title, user_id FROM tasks WHERE id = $1", task.ID,
	).Scan(&storedTitle, &storedUserID)
	if err != nil {
		t.Fatalf("failed to query created task: %v", err)
	}
	if storedTitle != "Created task" {
		t.Errorf("expected stored title %q, got %q", "Created task", storedTitle)
	}
	if storedUserID != userID {
		t.Errorf("expected stored user_id %d, got %d", userID, storedUserID)
	}
}

func TestPostTaskHandler_NoToken(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	rec := doRequest(t, router, http.MethodPost, "/tasks", `{"title":"No token"}`, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Errorf("expected unauthorized message, got body=%s", rec.Body.String())
	}
}

func TestPostTaskHandler_EmptyTitle(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-post-empty@example.com")
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodPost, "/tasks", `{"title":"   "}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Title cannot be empty") {
		t.Errorf("expected empty-title message, got body=%s", rec.Body.String())
	}
}

func TestPutTaskHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-put@example.com")
	taskID := insertTestTask(t, db, "Old title", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodPut, fmt.Sprintf("/tasks/%d", taskID), `{"title":"Updated title"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var task models.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if task.ID != taskID {
		t.Errorf("expected ID %d, got %d", taskID, task.ID)
	}
	if task.Title != "Updated title" {
		t.Errorf("expected title %q, got %q", "Updated title", task.Title)
	}

	var storedTitle string
	err := db.QueryRow("SELECT title FROM tasks WHERE id = $1", taskID).Scan(&storedTitle)
	if err != nil {
		t.Fatalf("failed to query updated task: %v", err)
	}
	if storedTitle != "Updated title" {
		t.Errorf("expected stored title %q, got %q", "Updated title", storedTitle)
	}
}

func TestPutTaskHandler_NoToken(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	rec := doRequest(t, router, http.MethodPut, "/tasks/1", `{"title":"No token"}`, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Errorf("expected unauthorized message, got body=%s", rec.Body.String())
	}
}

func TestPutTaskHandler_WrongOwner(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	owner := createTestUser(t, db, "ht-put-owner@example.com")
	intruder := createTestUser(t, db, "ht-put-intruder@example.com")
	taskID := insertTestTask(t, db, "Owner's title", owner)
	token := testToken(t, intruder)

	rec := doRequest(t, router, http.MethodPut, fmt.Sprintf("/tasks/%d", taskID), `{"title":"Hacked"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var storedTitle string
	err := db.QueryRow("SELECT title FROM tasks WHERE id = $1", taskID).Scan(&storedTitle)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if storedTitle != "Owner's title" {
		t.Errorf("expected stored title to remain %q, got %q", "Owner's title", storedTitle)
	}
}

func TestPutTaskHandler_NotFound(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-put-nf@example.com")
	taskID := insertTestTask(t, db, "Existing", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodPut, fmt.Sprintf("/tasks/%d", taskID+100000), `{"title":"Nope"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}
}

func TestPutTaskHandler_EmptyTitle(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-put-empty@example.com")
	taskID := insertTestTask(t, db, "Current", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodPut, fmt.Sprintf("/tasks/%d", taskID), `{"title":"   "}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Title cannot be empty") {
		t.Errorf("expected empty-title message, got body=%s", rec.Body.String())
	}
}

func TestDeleteTaskHandler_Success(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-delete@example.com")
	taskID := insertTestTask(t, db, "To delete", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/tasks/%d", taskID), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Task Deleted") {
		t.Errorf("expected deletion message, got body=%s", rec.Body.String())
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count task: %v", err)
	}
	if count != 0 {
		t.Errorf("expected task to be deleted, found %d rows", count)
	}
}

func TestDeleteTaskHandler_NoToken(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	rec := doRequest(t, router, http.MethodDelete, "/tasks/1", "", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Unauthorized") {
		t.Errorf("expected unauthorized message, got body=%s", rec.Body.String())
	}
}

func TestDeleteTaskHandler_WrongOwner(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	owner := createTestUser(t, db, "ht-delete-owner@example.com")
	intruder := createTestUser(t, db, "ht-delete-intruder@example.com")
	taskID := insertTestTask(t, db, "Not yours", owner)
	token := testToken(t, intruder)

	rec := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/tasks/%d", taskID), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count task: %v", err)
	}
	if count != 1 {
		t.Errorf("expected task to still exist, found %d rows", count)
	}
}

func TestDeleteTaskHandler_NotFound(t *testing.T) {
	db := requireTestDB(t)
	router := newTestRouter(newTestHandler(db))

	userID := createTestUser(t, db, "ht-delete-nf@example.com")
	taskID := insertTestTask(t, db, "Existing", userID)
	token := testToken(t, userID)

	rec := doRequest(t, router, http.MethodDelete, fmt.Sprintf("/tasks/%d", taskID+100000), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body=%s", rec.Code, rec.Body.String())
	}
}
