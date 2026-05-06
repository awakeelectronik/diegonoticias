package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/awakeelectronik/diegonoticias/internal/auth"
	"github.com/awakeelectronik/diegonoticias/internal/config"
	"golang.org/x/term"
)

func runSetupAdmin() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config inválida: %v\n", err)
		os.Exit(2)
	}
	adminPath := filepath.Join(cfg.DataDir, "admin.json")
	envUsername := strings.TrimSpace(os.Getenv("DN_SETUP_USERNAME"))
	envPassword := os.Getenv("DN_SETUP_PASSWORD")
	if envUsername != "" && envPassword != "" {
		writeAdminFile(adminPath, envUsername, envPassword)
		fmt.Printf("Admin creado en %s\n", adminPath)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	if _, err := os.Stat(adminPath); err == nil {
		fmt.Print("data/admin.json ya existe. ¿Sobrescribir? [s/N]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "s" && answer != "si" && answer != "sí" && answer != "y" {
			fmt.Println("Cancelado.")
			return
		}
	}

	fmt.Print("Username [diego]: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "diego"
	}

	fmt.Print("Password: ")
	pass1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo leer password: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("Repite password: ")
	pass2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo leer password: %v\n", err)
		os.Exit(1)
	}
	if string(pass1) != string(pass2) {
		fmt.Fprintln(os.Stderr, "las contraseñas no coinciden")
		os.Exit(1)
	}
	if len(strings.TrimSpace(string(pass1))) < 8 {
		fmt.Fprintln(os.Stderr, "la contraseña debe tener al menos 8 caracteres")
		os.Exit(1)
	}

	writeAdminFile(adminPath, username, string(pass1))
	fmt.Printf("Admin creado en %s\n", adminPath)
}

func writeAdminFile(adminPath, username, password string) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo crear hash: %v\n", err)
		os.Exit(1)
	}
	now := time.Now()
	a := auth.AdminUser{
		Username:     username,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := auth.SaveAdmin(adminPath, a); err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo guardar admin: %v\n", err)
		os.Exit(1)
	}
}

