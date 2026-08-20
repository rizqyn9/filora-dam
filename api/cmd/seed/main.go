package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"

	"github.com/rizqyn9/filora-dam/api/internal/config"
	"github.com/rizqyn9/filora-dam/api/internal/database"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Filora Superuser Setup ===")
	fmt.Println()

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Password (min 8 chars): ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	if err := queries.AssignRole(ctx, db.AssignRoleParams{UserID: user.ID, RoleName: "superuser"}); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	fmt.Println()
	fmt.Printf("Superuser created: %s (%s) — ID: %d\n", email, name, user.ID)
	fmt.Println("Done.")
	return nil
}
