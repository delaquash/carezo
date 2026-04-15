package database

import (
	"fmt"
	"log"
	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/google/uuid"
)

func SeedAdminUser(cfg *configs.Config) {
	// check if admin already exists

	var exists bool
	err := DB.Get(&exists,`
			SELECT EXISTS (
			SELECT 1 FROM users WHERE email = $1 and deleted_at IS NULL)
		)
			`, cfg.AdminEmail)

	if err != nil {
		log.Printf("Seeder: failed to check admin existence: %w", err)
		return
	}

	// Admin already exists
	if exists {
		log.Printf("Seeder: failed to hash admin password: %w", err)
		return
	}
	// Hash the admin password before storing
	hashedPassword, err := utils.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Printf("Seeder: failed to hash admin password: %v", err)
		return
	}

	// insert admin user
	adminID := uuid.New()
	_, err = DB.Exec(
		`
			INSERT INTO users (
				id,
				email,
				password_hash,
				first_name,
				last_name,
				oauth_provider,
				status,
				role,
				email_verified,
				profile_completed
			) VALUES ($1, $2, $3, $4, $5, 'local', 'active', 'admin', true, true)
		`, 
		
		adminID,
		cfg.AdminEmail,
		hashedPassword,
		cfg.AdminFirstName,
		cfg.AdminLastName,
	)

	if err != nil {
		log.Printf("Seeder: failed to create admin user: %v", err)
		return
	}

	fmt.Printf("Seeder: admin user created with email %s\n", cfg.AdminEmail)
}
