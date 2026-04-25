package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"conductor/internal/db"
	"conductor/internal/models"
)

func main() {
	dbPath := "conductor.db"
	if p := os.Getenv("CONDUCTOR_DB_PATH"); p != "" {
		dbPath = p
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := seedDB(database); err != nil {
		log.Fatalf("seed: %v", err)
	}

	fmt.Println("seed complete: 3 users, 5 projects, 18 tasks")
	fmt.Println("  login with any username and password 'password'")
}

func seedDB(d *db.DB) error {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}
	pw := string(hash)

	aliceID, err := d.CreateUser("alice", pw)
	if err != nil {
		return fmt.Errorf("create alice: %w", err)
	}
	bobID, err := d.CreateUser("bob", pw)
	if err != nil {
		return fmt.Errorf("create bob: %w", err)
	}
	charlieID, err := d.CreateUser("charlie", pw)
	if err != nil {
		return fmt.Errorf("create charlie: %w", err)
	}

	renovationID, err := d.CreateProject("Home Renovation", "Phase 1: kitchen and bathrooms")
	if err != nil {
		return fmt.Errorf("create renovation project: %w", err)
	}
	groceriesID, err := d.CreateProject("Weekly Groceries", "")
	if err != nil {
		return fmt.Errorf("create groceries project: %w", err)
	}
	gardenID, err := d.CreateProject("Garden", "Spring planting and maintenance")
	if err != nil {
		return fmt.Errorf("create garden project: %w", err)
	}
	carID, err := d.CreateProject("Car Maintenance", "")
	if err != nil {
		return fmt.Errorf("create car project: %w", err)
	}
	if _, err = d.CreateProject("Empty Project", "No tasks yet — test empty state"); err != nil {
		return fmt.Errorf("create empty project: %w", err)
	}

	overdue := time.Now().AddDate(0, 0, -7)
	thisWeek := time.Now().AddDate(0, 0, 3)
	later := time.Now().AddDate(0, 0, 30)
	past := time.Now().AddDate(0, 0, -14)

	tasks := []db.CreateTaskParams{
		// Home Renovation — 6 tasks
		{
			Title:       "Fix roof leak",
			Description: "Water damage visible on ceiling in master bedroom. Needs contractor assessment.",
			Status:      models.StatusBlocked,
			Priority:    models.PriorityCritical,
			Category:    "maintenance",
			ProjectID:   renovationID,
			AssigneeID:  &aliceID,
			DueDate:     &overdue,
			CreatedBy:   aliceID,
		},
		{
			Title:      "Replace kitchen faucet",
			Status:     models.StatusInProgress,
			Priority:   models.PriorityHigh,
			Category:   "repairs",
			ProjectID:  renovationID,
			AssigneeID: &bobID,
			DueDate:    &thisWeek,
			CreatedBy:  aliceID,
		},
		{
			Title:     "Paint living room",
			Status:    models.StatusTodo,
			Priority:  models.PriorityMedium,
			Category:  "painting",
			ProjectID: renovationID,
			DueDate:   &later,
			CreatedBy: bobID,
		},
		{
			Title:      "Install new shelving",
			Status:     models.StatusDone,
			Priority:   models.PriorityLow,
			Category:   "installations",
			ProjectID:  renovationID,
			AssigneeID: &aliceID,
			DueDate:    &past,
			CreatedBy:  aliceID,
		},
		{
			Title:      "Order bathroom tiles",
			Link:       "https://example.com/tiles",
			Status:     models.StatusTodo,
			Priority:   models.PriorityHigh,
			Category:   "shopping",
			ProjectID:  renovationID,
			AssigneeID: &charlieID,
			DueDate:    &thisWeek,
			CreatedBy:  aliceID,
		},
		{
			Title:      "Repair front steps",
			Status:     models.StatusTodo,
			Priority:   models.PriorityMedium,
			Category:   "repairs",
			ProjectID:  renovationID,
			AssigneeID: &bobID,
			CreatedBy:  bobID,
		},
		// Weekly Groceries — 3 tasks
		{
			Title:      "Buy vegetables",
			Status:     models.StatusTodo,
			Priority:   models.PriorityMedium,
			Category:   "shopping",
			ProjectID:  groceriesID,
			AssigneeID: &aliceID,
			DueDate:    &thisWeek,
			CreatedBy:  aliceID,
		},
		{
			Title:      "Restock pantry",
			Status:     models.StatusTodo,
			Priority:   models.PriorityLow,
			Category:   "shopping",
			ProjectID:  groceriesID,
			AssigneeID: &bobID,
			DueDate:    &thisWeek,
			CreatedBy:  aliceID,
		},
		{
			Title:      "Get cleaning supplies",
			Status:     models.StatusDone,
			Priority:   models.PriorityLow,
			Category:   "shopping",
			ProjectID:  groceriesID,
			AssigneeID: &aliceID,
			DueDate:    &past,
			CreatedBy:  bobID,
		},
		// Garden — 5 tasks
		{
			Title:     "Fix irrigation system",
			Status:    models.StatusBlocked,
			Priority:  models.PriorityHigh,
			Category:  "maintenance",
			ProjectID: gardenID,
			DueDate:   &overdue,
			CreatedBy: aliceID,
		},
		{
			Title:      "Trim hedges",
			Status:     models.StatusInProgress,
			Priority:   models.PriorityLow,
			Category:   "garden",
			ProjectID:  gardenID,
			AssigneeID: &charlieID,
			DueDate:    &thisWeek,
			CreatedBy:  bobID,
		},
		{
			Title:      "Plant tomatoes",
			Status:     models.StatusTodo,
			Priority:   models.PriorityMedium,
			Category:   "garden",
			ProjectID:  gardenID,
			AssigneeID: &bobID,
			DueDate:    &later,
			CreatedBy:  aliceID,
		},
		{
			Title:     "Build raised bed",
			Status:    models.StatusTodo,
			Priority:  models.PriorityMedium,
			Category:  "garden",
			ProjectID: gardenID,
			CreatedBy: aliceID,
		},
		{
			Title:      "Order compost",
			Status:     models.StatusDone,
			Priority:   models.PriorityLow,
			Category:   "shopping",
			ProjectID:  gardenID,
			AssigneeID: &aliceID,
			DueDate:    &past,
			CreatedBy:  aliceID,
		},
		// Car Maintenance — 4 tasks
		{
			Title:     "Check brakes",
			Status:    models.StatusTodo,
			Priority:  models.PriorityCritical,
			Category:  "maintenance",
			ProjectID: carID,
			DueDate:   &thisWeek,
			CreatedBy: bobID,
		},
		{
			Title:      "Oil change",
			Status:     models.StatusDone,
			Priority:   models.PriorityHigh,
			Category:   "maintenance",
			ProjectID:  carID,
			AssigneeID: &aliceID,
			DueDate:    &past,
			CreatedBy:  aliceID,
		},
		{
			Title:      "Rotate tires",
			Status:     models.StatusDone,
			Priority:   models.PriorityMedium,
			Category:   "maintenance",
			ProjectID:  carID,
			AssigneeID: &bobID,
			DueDate:    &past,
			CreatedBy:  bobID,
		},
		{
			Title:      "Replace windshield wipers",
			Status:     models.StatusDone,
			Priority:   models.PriorityLow,
			Category:   "maintenance",
			ProjectID:  carID,
			AssigneeID: &aliceID,
			DueDate:    &past,
			CreatedBy:  aliceID,
		},
	}

	for _, p := range tasks {
		if _, err := d.CreateTask(p); err != nil {
			return fmt.Errorf("create task %q: %w", p.Title, err)
		}
	}

	return nil
}
