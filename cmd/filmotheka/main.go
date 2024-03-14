package main

import (
	"log"
	"net/http"

	"github.com/axywe/filmotheka/pkg/actor"
	"github.com/axywe/filmotheka/pkg/movie"
	"github.com/axywe/filmotheka/pkg/storage"
)

func main() {
	db, err := storage.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	actorHandler := actor.NewHandler(db)
	movieHandler := movie.NewHandler(db)

	http.HandleFunc("/actors", actorHandler.ServeHTTP)
	http.HandleFunc("/movies", movieHandler.ServeHTTP)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
