package main

import (
	"log"
	"net/http"

	"github.com/axywe/filmotheka_vk/internal/auth"
	"github.com/axywe/filmotheka_vk/internal/middleware"
	"github.com/axywe/filmotheka_vk/pkg/actor"
	"github.com/axywe/filmotheka_vk/pkg/movie"
	"github.com/axywe/filmotheka_vk/pkg/storage"
)

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	actorHandler := actor.NewHandler(db)
	movieHandler := movie.NewHandler(db)
	authHandler := auth.NewHandler(db)

	http.Handle("/actors", middleware.RoleCheckMiddleware(actorHandler))
	http.Handle("/movies", middleware.RoleCheckMiddleware(movieHandler))
	http.HandleFunc("/auth", authHandler.ServeHTTP)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
