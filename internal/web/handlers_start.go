package web

import (
	"net/http"
	"strings"

	"adventure/internal/game"
)

const maxNameLen = 64

func allowedAvatar(id string) bool {
	for _, a := range AvatarOptions {
		if a == id {
			return true
		}
	}
	return false
}

// adventureOptions builds AdventureOption list from Engine.Stories (ID + Title or derived name).
func (s *Server) adventureOptions() []AdventureOption {
	if s.Engine == nil || s.Engine.Stories == nil {
		return nil
	}
	out := make([]AdventureOption, 0, len(s.Engine.Stories))
	for id, story := range s.Engine.Stories {
		name := story.Title
		if name == "" && id != "" {
			name = strings.ToUpper(id[:1]) + id[1:]
		}
		out = append(out, AdventureOption{ID: id, Name: name})
	}
	return out
}

func (s *Server) defaultStoryID() string {
	if s.Engine == nil || s.Engine.Stories == nil {
		return game.DefaultStoryID
	}
	if s.Engine.Stories[game.DefaultStoryID] != nil {
		return game.DefaultStoryID
	}
	for id := range s.Engine.Stories {
		return id
	}
	return game.DefaultStoryID
}

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
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	defaultID := s.defaultStoryID()
	defaultStory := s.Engine.Stories[defaultID]
	if defaultStory == nil {
		http.Error(w, "no adventure available", 500)
		return
	}
	st := game.NewPlayer(defaultID, defaultStory.Start)
	stats, statDice := game.RollStatsDetailed()
	st.Stats = stats

	if err := s.Store.Put(ctx, id, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{
		Stats:            st.Stats,
		StrengthDice:     statDice[0],
		LuckDice:         statDice[1],
		HealthDice:       statDice[2],
		SessionID:        id,
		Name:             st.Name,
		Avatar:           st.Avatar,
		AvatarOptions:    AvatarOptions,
		StoryID:          st.StoryID,
		AdventureOptions: s.adventureOptions(),
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	// Preserve current name and avatar from the form so reroll doesn't reset character selection
	name := strings.TrimSpace(r.FormValue("name"))
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	st.Name = name
	avatar := r.FormValue("avatar")
	if allowedAvatar(avatar) {
		st.Avatar = avatar
	}
	storyID := r.FormValue("story_id")
	if s.Engine != nil && s.Engine.Stories != nil && s.Engine.Stories[storyID] != nil {
		st.StoryID = storyID
	}

	stats, statDice := game.RollStatsDetailed()
	st.Stats = stats
	if err := s.Store.Put(ctx, sessionID, st); err != nil {
		http.Error(w, "failed to save state", 500)
		return
	}

	vm := StartViewModel{
		Stats:            st.Stats,
		StrengthDice:     statDice[0],
		LuckDice:         statDice[1],
		HealthDice:       statDice[2],
		SessionID:        sessionID,
		Name:             st.Name,
		Avatar:           st.Avatar,
		AvatarOptions:    AvatarOptions,
		StoryID:          st.StoryID,
		AdventureOptions: s.adventureOptions(),
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
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			})
			storyID := r.FormValue("story_id")
			if s.Engine.Stories[storyID] != nil {
				st.StoryID = storyID
				st.NodeID = s.Engine.Stories[storyID].Start
			} else {
				defaultID := s.defaultStoryID()
				if s.Engine.Stories[defaultID] != nil {
					st.StoryID = defaultID
					st.NodeID = s.Engine.Stories[defaultID].Start
				}
			}
			name := strings.TrimSpace(r.FormValue("name"))
			if len(name) > maxNameLen {
				name = name[:maxNameLen]
			}
			st.Name = name
			avatar := r.FormValue("avatar")
			if !allowedAvatar(avatar) {
				avatar = game.DefaultAvatar
			}
			st.Avatar = avatar
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

	storyID := r.FormValue("story_id")
	if s.Engine.Stories[storyID] != nil {
		st.StoryID = storyID
		st.NodeID = s.Engine.Stories[storyID].Start
	} else {
		defaultID := s.defaultStoryID()
		if s.Engine.Stories[defaultID] != nil {
			st.StoryID = defaultID
			st.NodeID = s.Engine.Stories[defaultID].Start
		}
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	st.Name = name
	avatar := r.FormValue("avatar")
	if !allowedAvatar(avatar) {
		avatar = game.DefaultAvatar
	}
	st.Avatar = avatar
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
