package repository

import (
	"database/sql"
	"errors"
	"task-manager-api/apperrors"
	"testing"
)

func TestUserRepo_RegisterUser_Success(t *testing.T) {
	db := requireTestDB(t)

	repo := &UserRepo{DB: db}
	email := "register@example.com"
	password := "hashed-password"

	err := repo.RegisterUser(email, password)
	if err != nil {
		t.Fatalf("RegisterUser returned error: %v", err)
	}

	var storedEmail, storedPassword string
	err = db.QueryRow(
		"SELECT email, password FROM users WHERE email = $1", email,
	).Scan(&storedEmail, &storedPassword)
	if err != nil {
		t.Fatalf("failed to query registered user: %v", err)
	}
	if storedEmail != email {
		t.Errorf("expected email %q, got %q", email, storedEmail)
	}
	if storedPassword != password {
		t.Errorf("expected password %q, got %q", password, storedPassword)
	}
}

func TestUserRepo_RegisterUser_DuplicateEmail(t *testing.T) {
	db := requireTestDB(t)

	repo := &UserRepo{DB: db}
	email := "duplicate@example.com"

	if err := repo.RegisterUser(email, "hashed-password"); err != nil {
		t.Fatalf("first RegisterUser returned error: %v", err)
	}

	err := repo.RegisterUser(email, "hashed-password")
	if !errors.Is(err, apperrors.ExistEmailError) {
		t.Fatalf("expected ExistEmailError, got %v", err)
	}
}

func TestUserRepo_GetUserByEmail_Success(t *testing.T) {
	db := requireTestDB(t)

	userID := createTestUser(t, db, "getuser@example.com")

	repo := &UserRepo{DB: db}
	user, err := repo.GetUserByEmail("getuser@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail returned error: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected ID %d, got %d", userID, user.ID)
	}
	if user.Email != "getuser@example.com" {
		t.Errorf("expected email %q, got %q", "getuser@example.com", user.Email)
	}
	if user.Password != "hashed-password" {
		t.Errorf("expected password %q, got %q", "hashed-password", user.Password)
	}
	if user.Role != "user" {
		t.Errorf("expected role %q, got %q", "user", user.Role)
	}
}

func TestUserRepo_GetUserByEmail_NotFound(t *testing.T) {
	db := requireTestDB(t)

	repo := &UserRepo{DB: db}
	_, err := repo.GetUserByEmail("nosuchuser@example.com")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
