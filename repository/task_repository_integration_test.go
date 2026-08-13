package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"task-manager-api/apperrors"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION") == "1" {
		if err := godotenv.Load("../.env"); err != nil {
			fmt.Println("failed to load .env file:", err)
			os.Exit(1)
		}
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
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=taskmanager_test sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword,
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

func TestTaskRepo_GetTasksByUser(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "integration@example.com")

	for _, title := range []string{"First task", "Second task"} {
		_, err := db.Exec("INSERT INTO tasks (title, user_id) VALUES ($1, $2)", title, userID)
		if err != nil {
			t.Fatalf("failed to insert test task: %v", err)
		}
	}

	repo := &TaskRepository{DB: db}
	tasks, err := repo.GetTasksByUser(userID)
	if err != nil {
		t.Fatalf("GetTasksByUser returned error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	titles := make(map[string]bool, len(tasks))
	for _, task := range tasks {
		if task.ID == 0 {
			t.Errorf("expected non-zero task ID, got 0")
		}
		titles[task.Title] = true
	}
	if !titles["First task"] || !titles["Second task"] {
		t.Errorf("expected both task titles, got %v", tasks)
	}
}

func TestTaskRepo_CreateTask(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "create@example.com")

	repo := &TaskRepository{DB: db}
	task, err := repo.CreateTask("New task", userID)
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if task.ID == 0 {
		t.Errorf("expected non-zero task ID, got 0")
	}
	if task.Title != "New task" {
		t.Errorf("expected title %q, got %q", "New task", task.Title)
	}

	var storedTitle string
	var storedUserID int
	err = db.QueryRow("SELECT title, user_id FROM tasks WHERE id = $1", task.ID).Scan(&storedTitle, &storedUserID)
	if err != nil {
		t.Fatalf("failed to query created task: %v", err)
	}
	if storedTitle != "New task" {
		t.Errorf("expected stored title %q, got %q", "New task", storedTitle)
	}
	if storedUserID != userID {
		t.Errorf("expected stored user_id %d, got %d", userID, storedUserID)
	}
}

func TestTaskRepo_GetTaskByID_Success(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "getbyid@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"My task", userID,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	task, err := repo.GetTaskByID(taskID, userID)
	if err != nil {
		t.Fatalf("GetTaskByID returned error: %v", err)
	}
	if task.ID != taskID {
		t.Errorf("expected ID %d, got %d", taskID, task.ID)
	}
	if task.Title != "My task" {
		t.Errorf("expected title %q, got %q", "My task", task.Title)
	}
}

func TestTaskRepo_GetTaskByID_WrongOwner(t *testing.T) {
	db := requireTestDB(t)

	userA := createTestUser(t, db, "getownerA@example.com")
	userB := createTestUser(t, db, "getownerB@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"Secret task", userA,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	_, err = repo.GetTaskByID(taskID, userB)
	if !errors.Is(err, apperrors.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskRepo_GetTaskByID_NotFound(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "getnotfound@example.com")

	repo := &TaskRepository{DB: db}
	_, err := repo.GetTaskByID(999999, userID)
	if !errors.Is(err, apperrors.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskRepo_UpdateTask_Success(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "update@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"Old title", userID,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	updated, err := repo.UpdateTask("New title", taskID, userID)
	if err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	if !updated {
		t.Errorf("expected UpdateTask to return true")
	}

	var storedTitle string
	err = db.QueryRow("SELECT title FROM tasks WHERE id = $1", taskID).Scan(&storedTitle)
	if err != nil {
		t.Fatalf("failed to query updated task: %v", err)
	}
	if storedTitle != "New title" {
		t.Errorf("expected stored title %q, got %q", "New title", storedTitle)
	}
}

func TestTaskRepo_UpdateTask_WrongOwner(t *testing.T) {
	db := requireTestDB(t)

	userA := createTestUser(t, db, "updateownerA@example.com")
	userB := createTestUser(t, db, "updateownerB@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"Original title", userA,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	updated, err := repo.UpdateTask("Hacked title", taskID, userB)
	if err != nil {
		t.Fatalf("UpdateTask returned error: %v", err)
	}
	if updated {
		t.Errorf("expected UpdateTask to return false for non-owner")
	}

	var storedTitle string
	err = db.QueryRow("SELECT title FROM tasks WHERE id = $1", taskID).Scan(&storedTitle)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if storedTitle != "Original title" {
		t.Errorf("expected stored title to remain %q, got %q", "Original title", storedTitle)
	}
}

func TestTaskRepo_DeleteTask_Success(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "delete@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"To delete", userID,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	deleted, err := repo.DeleteTask(taskID, userID)
	if err != nil {
		t.Fatalf("DeleteTask returned error: %v", err)
	}
	if !deleted {
		t.Errorf("expected DeleteTask to return true")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count task: %v", err)
	}
	if count != 0 {
		t.Errorf("expected task to be deleted, found %d rows", count)
	}
}

func TestTaskRepo_DeleteTask_WrongOwner(t *testing.T) {
	db := requireTestDB(t)

	userA := createTestUser(t, db, "deleteownerA@example.com")
	userB := createTestUser(t, db, "deleteownerB@example.com")

	var taskID int
	err := db.QueryRow(
		"INSERT INTO tasks (title, user_id) VALUES ($1, $2) RETURNING id",
		"Not yours", userA,
	).Scan(&taskID)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	repo := &TaskRepository{DB: db}
	deleted, err := repo.DeleteTask(taskID, userB)
	if err != nil {
		t.Fatalf("DeleteTask returned error: %v", err)
	}
	if deleted {
		t.Errorf("expected DeleteTask to return false for non-owner")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = $1", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count task: %v", err)
	}
	if count != 1 {
		t.Errorf("expected task to still exist, found %d rows", count)
	}
}
