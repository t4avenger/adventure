package web

import (
	"net/http"

	"adventure/internal/game"
)

// GET /start
func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	}

	st := game.NewPlayer(s.Engine.Story.Start)
	st.Stats = game.RollStats()

	if err := s.Store.Put(ctx, id, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{
		Stats: st.Stats,
	}

	// IMPORTANT: render layout, but tell it to use start.html
	if err := s.Tmpl.ExecuteTemplate(w, "layout.html", map[string]any{
		"Start": vm,
	}); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}

// POST /reroll
func (s *Server) handleReroll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, sessionID := s.getOrCreateState(ctx, w, r)

	st.Stats = game.RollStats()
	if err := s.Store.Put(ctx, sessionID, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{Stats: st.Stats}
	if err := s.Tmpl.ExecuteTemplate(w, "start.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}

// POST /begin
func (s *Server) handleBegin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st, sessionID := s.getOrCreateState(ctx, w, r)

	st.NodeID = s.Engine.Story.Start
	if err := s.Store.Put(ctx, sessionID, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm, err := s.makeViewModel(&st, "", nil, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := s.Tmpl.ExecuteTemplate(w, "game.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}
