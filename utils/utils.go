/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
)

func GetEnvVar(variable string, redact bool) (string, error) {
	v := os.Getenv(variable)
	if v == "" {
		err := errors.New("variable " + variable + " is empty. exiting")
		return "", err
	}

	if redact {
		fmt.Printf("Set %v to <redacted>\n", variable)
	} else {
		fmt.Printf("Set %v to %v\n", variable, v)
	}

	return v, nil
}
