// Package web provides HTTP handlers for the adventure game web interface.
package web

import (
	"context"
	"html/template"
	"net/http"
	"strconv"

	"adventure/internal/game"
	"adventure/internal/session"
)

// Server handles HTTP requests for the adventure game.
type Server struct {
	Engine     *game.Engine
	Store      session.Store[game.PlayerState]
	Tmpl       *template.Template
	StoriesDir string // optional; base dir for stories (scenery handler; tests set to temp dir)
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
	mux.HandleFunc("/map", s.handleMap)
	mux.HandleFunc("/scenery/", s.handleScenery)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/start", http.StatusFound)
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
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
	vm, err := s.makeViewModel(&res.State, msg, res.LastRoll, res.LastOutcome, res.LastPlayerDice, res.LastEnemyDice)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// htmx: return #game fragment + OOB sidebars; client skips sync and only runs dice animation
	w.Header().Set("X-Adventure-OOB", "true")
	if err := s.Tmpl.ExecuteTemplate(w, "game_response.html", vm); err != nil {
		http.Error(w, "failed to render template", 500)
		return
	}
}

func (s *Server) getOrCreateState(ctx context.Context, w http.ResponseWriter, r *http.Request) (state game.PlayerState, sessionID string, found bool) {
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
		defaultStoryID := game.DefaultStoryID
		defaultStory := s.Engine.Stories[defaultStoryID]
		if defaultStory == nil {
			// No default story; use first available
			for id, st := range s.Engine.Stories {
				defaultStoryID = id
				defaultStory = st
				break
			}
		}
		if defaultStory != nil {
			state = game.NewPlayer(defaultStoryID, defaultStory.Start)
		} else {
			state = game.NewPlayer("", "")
		}
		_ = s.Store.Put(ctx, id, state) //nolint:errcheck // Best effort: continue even if store fails
		return state, id, true
	}

	var ok bool
	var err error
	state, ok, err = s.Store.Get(ctx, id)
	if err != nil || !ok {
		// Session exists but state not found (e.g. store cleared). Caller should redirect to /start.
		return game.PlayerState{}, id, false
	}
	return state, id, true
}

func (s *Server) sessionID(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// BattleChoice is a single choice for battle (attack/luck on target or run).
type BattleChoice struct {
	Key  string
	Text string
}

// ViewModel contains data for rendering a game view.
type ViewModel struct {
	Node               *game.Node
	State              game.PlayerState
	Message            string
	LastRoll           *int
	LastPlayerDice     *[2]int
	LastEnemyDice      *[2]int
	LastOutcome        *string
	Enemies            []game.EnemyState // 1–3 or single horde for display
	BattleChoicePrefix string            // e.g. "battle" for keys battle:attack:0
	EffectiveChoices   []BattleChoice    // when in battle, synthetic choices; else nil
}

func (s *Server) makeViewModel(st *game.PlayerState, msg string, roll *int, outcome *string, playerDice, enemyDice *[2]int) (ViewModel, error) {
	n, err := s.Engine.CurrentNode(st)
	if err != nil {
		return ViewModel{}, err
	}
	vm := ViewModel{
		Node:           n,
		State:          *st,
		Message:        msg,
		LastRoll:       roll,
		LastPlayerDice: playerDice,
		LastEnemyDice:  enemyDice,
		LastOutcome:    outcome,
		Enemies:        st.Enemies,
	}
	if len(st.Enemies) > 0 {
		var battleKey string
		for i := range n.Choices {
			if n.Choices[i].Battle != nil {
				battleKey = n.Choices[i].Key
				break
			}
		}
		if battleKey != "" {
			vm.BattleChoicePrefix = battleKey
			for i, e := range st.Enemies {
				idxStr := strconv.Itoa(i)
				vm.EffectiveChoices = append(vm.EffectiveChoices,
					BattleChoice{Key: battleKey + ":attack:" + idxStr, Text: "Attack " + e.Name},
					BattleChoice{Key: battleKey + ":luck:" + idxStr, Text: "Luck " + e.Name},
				)
			}
			vm.EffectiveChoices = append(vm.EffectiveChoices, BattleChoice{Key: battleKey + ":run", Text: "Run away"})
		}
	}
	return vm, nil
}
