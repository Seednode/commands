/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envFile := filepath.FromSlash(homeDirectory + "/.config/errwrapper/.env")
	err = godotenv.Load(envFile)
	if err != nil {
		return err
	}

	return nil
}

func GetEnvVar(variable string) string {
	v := os.Getenv(variable)
	if v == "" {
		fmt.Println("Variable " + variable + " is empty. Exiting.")
		os.Exit(1)
	}

	return v
}
