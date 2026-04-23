package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	_ "time/tzdata"

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
		resp, err := http.Get("http://localhost:8080/healthz")
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

// adminAddUser, adminDeleteUser, adminResetPassword are stubs — implemented fully in Task 9.
func adminAddUser(database *db.DB, args []string) {
	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
}

func adminDeleteUser(database *db.DB, args []string) {
	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
}

func adminResetPassword(database *db.DB, args []string) {
	fmt.Fprintln(os.Stderr, "not yet implemented")
	os.Exit(1)
}
