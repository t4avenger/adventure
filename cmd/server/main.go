// Package main provides the entry point for the adventure game server.
package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

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
		"templates/layout_head.html",
		"templates/sidebar_left.html",
		"templates/sidebar_right.html",
		"templates/sidebar_left_oob.html",
		"templates/sidebar_right_oob.html",
		"templates/game.html",
		"templates/game_response.html",
		"templates/start.html",
	))

	srv := &web.Server{
		Engine: &game.Engine{Story: story},
		Store:  session.NewMemoryStore[game.PlayerState](),
		Tmpl:   tmpl,
	}

	s := &http.Server{
		Addr:         ":8080",
		Handler:      srv.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Println("listening on http://localhost:8080")
	log.Fatal(s.ListenAndServe())
}
