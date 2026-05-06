package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type AdminUser struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func LoadAdmin(path string) (*AdminUser, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read admin file: %w", err)
	}
	var a AdminUser
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("parse admin file: %w", err)
	}
	return &a, nil
}

func SaveAdmin(path string, admin AdminUser) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	b, err := json.MarshalIndent(admin, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal admin file: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write admin file: %w", err)
	}
	return nil
}

