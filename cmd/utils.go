/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
)

func GetEnvVar(variable, flag string, redact bool) (string, error) {
	var v string

	if flag != "" {
		v = flag
	} else {
		v = os.Getenv(variable)
	}

	if v == "" {
		err := errors.New(variable + " is empty. exiting")
		return "", err
	}

	if redact {
		fmt.Printf("Set %v to <redacted>\n", variable)
	} else {
		fmt.Printf("Set %v to %v\n", variable, v)
	}

	return v, nil
}
