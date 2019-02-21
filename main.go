package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fabiocampos/go-and-destroy/handlers"
	"github.com/fabiocampos/go-and-destroy/services"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/content/", http.StripPrefix("/content/", fs))

	gameService := services.NewGameService()
	http.HandleFunc("/game", handlers.GameHandler(gameService))
	go gameService.RunGame()

	http.ListenAndServe(":"+port, nil)
}
