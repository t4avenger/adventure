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

	_ = s.Store.Put(ctx, id, st)

	vm := StartViewModel{
		Stats: st.Stats,
	}

	// IMPORTANT: render layout, but tell it to use start.html
	_ = s.Tmpl.ExecuteTemplate(w, "layout.html", map[string]any{
		"Start": vm,
	})
}

// POST /reroll
func (s *Server) handleReroll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st := s.getOrCreateState(ctx, w, r)

	st.Stats = game.RollStats()
	_ = s.Store.Put(ctx, s.sessionID(r), st)

	vm := StartViewModel{Stats: st.Stats}
	_ = s.Tmpl.ExecuteTemplate(w, "start.html", vm)
}

// POST /begin
func (s *Server) handleBegin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	st := s.getOrCreateState(ctx, w, r)

	st.NodeID = s.Engine.Story.Start
	_ = s.Store.Put(ctx, s.sessionID(r), st)

	vm, err := s.makeViewModel(st, "", nil, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_ = s.Tmpl.ExecuteTemplate(w, "game.html", vm)
}
