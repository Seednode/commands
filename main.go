/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package main

import (
	"log"

	web "seedno.de/seednode/commands-web/web"
)

func main() {
	err := web.ServePage()
	if err != nil {
		log.Fatal(err)
	}
}
