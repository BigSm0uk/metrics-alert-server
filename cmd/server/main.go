package main

import (
	"log"
)

func main() {
	app, err := initializeApp()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
