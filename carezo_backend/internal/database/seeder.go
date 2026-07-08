package database

import (
	"log"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func SeedAdminUser(db *sqlx.DB, cfg *configs.Config) {
	var exists bool
	err := db.Get(&exists, ` 
        SELECT EXISTS (
            SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL
        )
    `, cfg.AdminEmail)

	if err != nil {
		log.Printf("Seeder: failed to check admin existence: %v", err)
		return
	}

	if exists {
		log.Println("Seeder: admin user already exists, skipping")
		return
	}

	hashedPassword, err := utils.HashPassword(cfg.AdminPassword)
	if err != nil {
		log.Printf("Seeder: failed to hash admin password: %v", err)
		return
	}

	adminID := uuid.New()
	_, err = db.Exec(`  
        INSERT INTO users (
            id, email, password_hash, first_name, last_name,
            oauth_provider, status, role, email_verified, profile_completed
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

	log.Printf("Seeder: admin user created with email %s\n", cfg.AdminEmail)
	log.Println("Admin Email:", cfg.AdminEmail)
	log.Println("Admin Password:", cfg.AdminPassword)
	log.Println("Admin FirstName:", cfg.AdminFirstName)
	log.Println("Admin LastName:", cfg.AdminLastName)
}
