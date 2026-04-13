package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/senran-N/sub2api/internal/config"
	"github.com/senran-N/sub2api/internal/repository"

	_ "github.com/lib/pq"
)

const (
	defaultMigrationTimeout = 10 * time.Minute
	defaultStatusTimeout    = 30 * time.Second
)

func main() {
	action := flag.String("action", "status", "migration action: up, down, status, plan-down")
	steps := flag.Int("steps", 1, "rollback steps for action=down")
	flag.Parse()

	cfg, err := config.LoadForBootstrap()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := openMigrationDB(cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	switch strings.ToLower(strings.TrimSpace(*action)) {
	case "up":
		runUp(db)
	case "down":
		runDown(db, *steps)
	case "status":
		runStatus(db)
	case "plan-down":
		runPlanDown(db, *steps)
	default:
		fmt.Fprintf(os.Stderr, "unsupported action %q, expected up|down|status|plan-down\n", *action)
		os.Exit(2)
	}
}

func openMigrationDB(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.Database.DSNWithTimezone(cfg.Timezone)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func runUp(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultMigrationTimeout)
	defer cancel()

	if err := repository.ApplyMigrations(ctx, db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("migrations applied")
}

func runDown(db *sql.DB, steps int) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultMigrationTimeout)
	defer cancel()

	if err := repository.RollbackMigrations(ctx, db, steps); err != nil {
		log.Fatalf("rollback migrations: %v", err)
	}
	log.Printf("rolled back %d migration(s)", steps)
}

func runStatus(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStatusTimeout)
	defer cancel()

	applied, err := repository.ListAppliedMigrations(ctx, db)
	if err != nil {
		log.Fatalf("list applied migrations: %v", err)
	}
	if len(applied) == 0 {
		log.Println("no applied migrations")
		return
	}

	for _, name := range applied {
		status, err := repository.GetRollbackMigrationStatus(name)
		if err != nil {
			log.Fatalf("get rollback status for %s: %v", name, err)
		}
		fmt.Printf("%s\tdown=%s\n", name, status)
	}
}

func runPlanDown(db *sql.DB, steps int) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStatusTimeout)
	defer cancel()

	plan, err := repository.PlanRollbackMigrations(ctx, db, steps)
	if err != nil {
		log.Fatalf("plan rollback migrations: %v", err)
	}
	if len(plan) == 0 {
		log.Println("no applied migrations to roll back")
		return
	}

	for i, item := range plan {
		fmt.Printf("%d\t%s\tdown=%s\n", i+1, item.UpName, item.Status)
	}
}
