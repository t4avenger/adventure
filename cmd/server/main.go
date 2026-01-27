package main

import (
	"html/template"
	"log"
	"net/http"

	"adventure/internal/game"
	"adventure/internal/session"
	"adventure/internal/web"
)

func main() {
	story, err := game.LoadStory("stories/demo.yaml")
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/layout.html",
		"templates/game.html",
		"templates/start.html",
	))

	srv := &web.Server{
		Engine: &game.Engine{Story: story},
		Store:  session.NewMemoryStore[game.PlayerState](),
		Tmpl:   tmpl,
	}

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Routes()))
}
