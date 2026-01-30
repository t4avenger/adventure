package web

import (
	"net/http"

	"adventure/internal/game"
)

// GET /start
func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Prevent caching so the user always sees the stats we just saved
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

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
	stats, statDice := game.RollStatsDetailed()
	st.Stats = stats

	if err := s.Store.Put(ctx, id, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{
		Stats:        st.Stats,
		StrengthDice: statDice[0],
		LuckDice:     statDice[1],
		HealthDice:   statDice[2],
		SessionID:    id,
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
	st, sessionID, found := s.getOrCreateState(ctx, w, r)
	if !found {
		http.Redirect(w, r, "/start", http.StatusFound)
		return
	}

	stats, statDice := game.RollStatsDetailed()
	st.Stats = stats
	if err := s.Store.Put(ctx, sessionID, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{
		Stats:        st.Stats,
		StrengthDice: statDice[0],
		LuckDice:     statDice[1],
		HealthDice:   statDice[2],
		SessionID:    sessionID,
	}
	if err := s.Tmpl.ExecuteTemplate(w, "start.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}

// POST /begin
func (s *Server) handleBegin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	// Prefer session_id from the start page so we always load the session that has the stats just shown
	sessionIDFromForm := r.FormValue("session_id")
	if sessionIDFromForm != "" {
		st, ok, err := s.Store.Get(ctx, sessionIDFromForm)
		if err == nil && ok {
			// Set cookie so future requests have it
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    sessionIDFromForm,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			st.NodeID = s.Engine.Story.Start
			if err := s.Store.Put(ctx, sessionIDFromForm, st); err != nil {
				http.Error(w, "failed to save state", 500)
				return
			}
			vm, err := s.makeViewModel(&st, "", nil, nil, nil, nil)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("X-Adventure-OOB", "true")
			if err := s.Tmpl.ExecuteTemplate(w, "game_response.html", vm); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			return
		}
	}

	st, sessionID, found := s.getOrCreateState(ctx, w, r)
	if !found {
		http.Redirect(w, r, "/start", http.StatusFound)
		return
	}

	st.NodeID = s.Engine.Story.Start
	if err := s.Store.Put(ctx, sessionID, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm, err := s.makeViewModel(&st, "", nil, nil, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("X-Adventure-OOB", "true")
	if err := s.Tmpl.ExecuteTemplate(w, "game_response.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}
