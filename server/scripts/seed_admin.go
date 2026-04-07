// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sovereign-fund/sovereign/config"
	"github.com/sovereign-fund/sovereign/internal/modules/admin/model"
	"github.com/sovereign-fund/sovereign/internal/shared/database"
	"github.com/sovereign-fund/sovereign/pkg/crypto"
	"github.com/sovereign-fund/sovereign/pkg/logger"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("usage: go run scripts/seed_admin.go <email> <password> <name>")
	}

	email := os.Args[1]
	password := os.Args[2]
	name := os.Args[3]

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config:", err)
	}

	l := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.NewPostgres(cfg.Database, l)
	if err != nil {
		log.Fatal("connect db:", err)
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		log.Fatal("hash password:", err)
	}

	admin := &model.AdminUser{
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		Role:         model.RoleSuperAdmin,
		IsActive:     true,
	}

	if err := db.Create(admin).Error; err != nil {
		log.Fatal("create admin:", err)
	}

	fmt.Printf("Super admin created: id=%s email=%s name=%s\n", admin.ID, admin.Email, admin.Name)
}
