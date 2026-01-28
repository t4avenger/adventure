// Package web provides HTTP handlers for the adventure game web interface.
package web

import (
	"context"
	"html/template"
	"net/http"

	"adventure/internal/game"
	"adventure/internal/session"
)

// Server handles HTTP requests for the adventure game.
type Server struct {
	Engine *game.Engine
	Store  session.Store[game.PlayerState]
	Tmpl   *template.Template
}

const cookieName = "adventure_sid"

// Routes returns an HTTP handler with all registered routes.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)

	mux.HandleFunc("/start", s.handleStart)
	mux.HandleFunc("/reroll", s.handleReroll)
	mux.HandleFunc("/begin", s.handleBegin)

	mux.HandleFunc("/play", s.handlePlay)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/start", http.StatusFound)
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, sessionID := s.getOrCreateState(ctx, w, r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	choice := r.FormValue("choice")

	res, err := s.Engine.ApplyChoice(&st, choice)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := s.Store.Put(ctx, sessionID, res.State); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	msg := res.ErrorMessage
	vm, err := s.makeViewModel(&res.State, msg, res.LastRoll, res.LastOutcome)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// htmx: return fragment for #game only
	if err := s.Tmpl.ExecuteTemplate(w, "game.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}

func (s *Server) getOrCreateState(ctx context.Context, w http.ResponseWriter, r *http.Request) (state game.PlayerState, sessionID string) {
	id := s.sessionID(r)
	if id == "" {
		id = s.Store.NewID()
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    id,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		state = game.NewPlayer(s.Engine.Story.Start)
		_ = s.Store.Put(ctx, id, state) //nolint:errcheck // Best effort: continue even if store fails
		return state, id
	}

	var ok bool
	var err error
	state, ok, err = s.Store.Get(ctx, id)
	if err != nil || !ok {
		// Create new state on store error or missing session
		state = game.NewPlayer(s.Engine.Story.Start)
		_ = s.Store.Put(ctx, id, state) //nolint:errcheck // Best effort: continue even if store fails
	}
	return state, id
}

func (s *Server) sessionID(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// ViewModel contains data for rendering a game view.
type ViewModel struct {
	Node        *game.Node
	State       game.PlayerState
	Message     string
	LastRoll    *int
	LastOutcome *string
}

func (s *Server) makeViewModel(st *game.PlayerState, msg string, roll *int, outcome *string) (ViewModel, error) {
	n, err := s.Engine.CurrentNode(st)
	if err != nil {
		return ViewModel{}, err
	}
	return ViewModel{
		Node:        n,
		State:       *st,
		Message:     msg,
		LastRoll:    roll,
		LastOutcome: outcome,
	}, nil
}
