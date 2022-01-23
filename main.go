package main

import (
	"log"
	"os"

	"github.com/cmars/oniongrok/app"
)

func main() {
	err := app.App().Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
