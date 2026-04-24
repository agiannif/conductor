package main

import (
	"testing"

	"conductor/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestAdminAddUser(t *testing.T) {
	database := openTestDB(t)

	adminAddUser(database, []string{"alice", "secret123"})

	user, hash, err := database.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if user.Username != "alice" {
		t.Errorf("got username %q, want %q", user.Username, "alice")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret123")); err != nil {
		t.Errorf("password hash does not match: %v", err)
	}
}

func TestAdminDeleteUser(t *testing.T) {
	database := openTestDB(t)

	// Create a user to delete.
	adminAddUser(database, []string{"alice", "secret123"})

	adminDeleteUser(database, []string{"alice"})

	_, _, err := database.GetUserByUsername("alice")
	if err == nil {
		t.Fatal("expected error after deleting user, got nil")
	}
}

func TestAdminResetPassword(t *testing.T) {
	database := openTestDB(t)

	// Create a user with old password.
	adminAddUser(database, []string{"alice", "oldpass"})

	adminResetPassword(database, []string{"alice", "newpass"})

	_, hash, err := database.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("newpass")); err != nil {
		t.Errorf("new password hash does not match: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("oldpass")); err == nil {
		t.Error("old password should not match the new hash")
	}
}
