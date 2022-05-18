/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
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

	envFile := filepath.FromSlash(homeDirectory + "/.config/commands/.env")
	err = godotenv.Load(envFile)
	if err != nil {
		return err
	}

	return nil
}

func GetEnvVar(variable string) (string, error) {
	v := os.Getenv(variable)
	if v == "" {
		err := errors.New("variable " + variable + " is empty. exiting")
		return "", err
	}

	fmt.Printf("Set %v to %v\n", variable, v)

	return v, nil
}
