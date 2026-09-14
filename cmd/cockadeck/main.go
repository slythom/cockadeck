package main

// point d'entrée de l'app avec la conf de router.go
import (
	"log"
	"net/http"

	"cockadeck/router"
)

func main() {
	mux := router.Setup()
	log.Fatal(http.ListenAndServe(":8080", mux))
}
