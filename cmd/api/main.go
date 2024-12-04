package main

import (
	"log"

	"github.com/drypb/api/internal/api"
)

func main() {
	err := api.Run()
	if err != nil {
		log.Fatal(err)
	}
}
