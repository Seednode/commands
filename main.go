/*
Copyright © 2023 Seednode <seednode@seedno.de>
*/

package main

import (
	"log"

	cmd "seedno.de/seednode/commands-web/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
