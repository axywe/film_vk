package main

import (
	"log"
	"net/http"

	_ "github.com/axywe/filmotheka_vk/docs"
	"github.com/axywe/filmotheka_vk/internal/auth"
	"github.com/axywe/filmotheka_vk/internal/middleware"
	"github.com/axywe/filmotheka_vk/pkg/actor"
	"github.com/axywe/filmotheka_vk/pkg/movie"
	"github.com/axywe/filmotheka_vk/pkg/storage"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Filmotheka API
// @description This is a server for Filmotheka application.
// @version 1.0
// @host 0.0.0.0:8080
// @BasePath /

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	actorHandler := actor.NewHandler(db)
	movieHandler := movie.NewHandler(db)
	tokenGenerator := &auth.JWTTokenGenerator{}
	authHandler := auth.NewHandler(db, tokenGenerator)

	http.Handle("/swagger/", httpSwagger.WrapHandler)
	http.Handle("/actors", middleware.RoleCheckMiddleware(actorHandler))
	http.Handle("/movies", middleware.RoleCheckMiddleware(movieHandler))

	http.HandleFunc("/auth", authHandler.ServeHTTP)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
