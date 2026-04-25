package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"golang.org/x/crypto/bcrypt"

	"conductor/internal/config"
	"conductor/internal/db"
	"conductor/internal/handlers"
	conductorweb "conductor/web"
)

func main() {
	if len(os.Args) > 1 {
		runCLI(os.Args[1:])
		return
	}

	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	h := handlers.New(database, cfg.SecureCookie, conductorweb.Static)
	fmt.Println("conductor listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", h))
}

func runCLI(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: conductor <command>")
		os.Exit(1)
	}
	switch args[0] {
	case "healthz":
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://localhost:8080/healthz")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	case "admin":
		runAdmin(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func runAdmin(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: conductor admin <command>")
		os.Exit(1)
	}
	cfg := config.Load()
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	switch args[0] {
	case "list-users":
		users, err := database.ListUsers()
		if err != nil {
			log.Fatalf("list users: %v", err)
		}
		for _, u := range users {
			fmt.Printf("%d\t%s\n", u.ID, u.Username)
		}
	case "add-user":
		adminAddUser(database, args[1:])
	case "delete-user":
		adminDeleteUser(database, args[1:])
	case "reset-password":
		adminResetPassword(database, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown admin command: %s\n", args[0])
		os.Exit(1)
	}
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}
	return string(hash)
}

func adminAddUser(database *db.DB, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: conductor admin add-user <username> <password>")
		os.Exit(1)
	}
	username, password := args[0], args[1]
	id, err := database.CreateUser(username, hashPassword(password))
	if err != nil {
		log.Fatalf("add user: %v", err)
	}
	fmt.Printf("created user %q (id=%d)\n", username, id)
}

func adminDeleteUser(database *db.DB, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: conductor admin delete-user <username>")
		os.Exit(1)
	}
	username := args[0]
	user, _, err := database.GetUserByUsername(username)
	if errors.Is(err, db.ErrNotFound) {
		fmt.Fprintf(os.Stderr, "user %q not found\n", username)
		os.Exit(1)
	} else if err != nil {
		log.Fatalf("get user: %v", err)
	}
	if err := database.DeleteUser(user.ID); err != nil {
		log.Fatalf("delete user: %v", err)
	}
	fmt.Printf("deleted user %q\n", username)
}

func adminResetPassword(database *db.DB, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: conductor admin reset-password <username> <new-password>")
		os.Exit(1)
	}
	username, newPassword := args[0], args[1]
	user, _, err := database.GetUserByUsername(username)
	if errors.Is(err, db.ErrNotFound) {
		fmt.Fprintf(os.Stderr, "user %q not found\n", username)
		os.Exit(1)
	} else if err != nil {
		log.Fatalf("get user: %v", err)
	}
	if err := database.UpdateUserPassword(user.ID, hashPassword(newPassword)); err != nil {
		log.Fatalf("reset password: %v", err)
	}
	fmt.Printf("password reset for user %q\n", username)
}
