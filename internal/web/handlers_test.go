package web

import (
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"adventure/internal/game"
	"adventure/internal/session"
)

const testStoryID = "test"
const testNodeRoad = "road"

func testServer(t *testing.T) *Server {
	t.Helper()
	story := &game.Story{
		Start: "start",
		Nodes: map[string]*game.Node{
			"start": {
				Text: "You are at the start.",
				Choices: []game.Choice{
					{Key: "next", Text: "Go next", Next: "end"},
				},
			},
			"end": {
				Text:   "The end.",
				Ending: true,
			},
		},
	}
	engine := &game.Engine{Stories: map[string]*game.Story{testStoryID: story}}
	store := session.NewMemoryStore[game.PlayerState]()

	tmplDir := filepath.Join("..", "..", "templates")
	tmpl := template.Must(template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "layout_head.html"),
		filepath.Join(tmplDir, "sidebar_left.html"),
		filepath.Join(tmplDir, "sidebar_right.html"),
		filepath.Join(tmplDir, "sidebar_left_oob.html"),
		filepath.Join(tmplDir, "sidebar_right_oob.html"),
		filepath.Join(tmplDir, "game.html"),
		filepath.Join(tmplDir, "game_response.html"),
		filepath.Join(tmplDir, "start.html"),
	))
	return &Server{Engine: engine, Store: store, Tmpl: tmpl}
}

const pathStart = "/start"

func TestHandleIndex(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("Expected 302, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != pathStart {
		t.Errorf("Expected Location %s, got %q", pathStart, loc)
	}
}

func TestHandleStart(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, pathStart, http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Create Your Adventurer") {
		t.Error("Expected body to contain 'Create Your Adventurer'")
	}
}

func TestHandleReroll(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 7, Luck: 7, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/reroll", strings.NewReader("session_id="+id+"&name=Hero&avatar=male_young"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	updated, ok, err := srv.Store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Expected updated session")
	}
	if !updated.RerollUsed {
		t.Error("Expected reroll flag set after first reroll")
	}
}

func TestHandleReroll_OncePerSession(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 99, Luck: 99, Health: 99}
	st.RerollUsed = true
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/reroll", strings.NewReader("session_id="+id+"&name=Hero&avatar=male_young"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	updated, ok, err := srv.Store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Expected updated session")
	}
	if updated.Stats != st.Stats {
		t.Errorf("Expected stats unchanged after reroll used, got %+v", updated.Stats)
	}
	if !updated.RerollUsed {
		t.Error("Expected reroll flag to stay set")
	}
}

func TestHandleBegin(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/begin", strings.NewReader("session_id="+id+"&name=Hero&avatar=female_young&story_id=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Adventure-OOB") != "true" {
		t.Error("Expected X-Adventure-OOB header")
	}
	updated, ok, err := srv.Store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Expected updated session")
	}
	if !updated.RerollUsed {
		t.Error("Expected reroll to be locked after begin")
	}
}

func TestHandlePlay(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=next"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "The end.") {
		t.Error("Expected body to contain 'The end.'")
	}
}

func TestHandlePlay_UnknownSessionRedirectsToStart(t *testing.T) {
	srv := testServer(t)
	// Cookie with ID that was never Put so Get returns not found -> redirect to /start
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=next"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "unknown-session-id-never-stored"})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("Expected 302, got %d", rec.Code)
	}
	if rec.Header().Get("Location") != pathStart {
		t.Errorf("Expected redirect to %s, got %q", pathStart, rec.Header().Get("Location"))
	}
}

func TestHandlePlay_NoCookie_CreatesStateAndRenders(t *testing.T) {
	srv := testServer(t)
	// No cookie: getOrCreateState creates new session and state from default story, then applies choice.
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=next"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "The end.") {
		t.Error("Expected body to contain 'The end.'")
	}
	// New session cookie should be set
	if rec.Header().Get("Set-Cookie") == "" {
		t.Error("Expected Set-Cookie for new session")
	}
}

func TestHandlePlay_EmptyStories_NewPlayerEmpty(t *testing.T) {
	// Server with no stories: getOrCreateState uses NewPlayer("", ""); CurrentNode then errors.
	srv := testServer(t)
	srv.Engine = &game.Engine{Stories: map[string]*game.Story{}}
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=next"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 when no stories, got %d", rec.Code)
	}
}

// errReader is an io.Reader that always returns an error (for testing ParseForm failure).
type errReader struct{ err error }

func (e *errReader) Read([]byte) (int, error) { return 0, e.err }

func TestHandlePlay_ParseFormError_BadRequest(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/play", io.NopCloser(&errReader{err: errors.New("read error")}))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 on ParseForm error, got %d", rec.Code)
	}
}

func TestAdventureOptions(t *testing.T) {
	srv := testServer(t)
	opts := srv.adventureOptions()
	if len(opts) != 1 {
		t.Errorf("Expected 1 adventure option, got %d", len(opts))
	}
	if len(opts) > 0 && opts[0].ID != testStoryID {
		t.Errorf("Expected ID %q, got %q", testStoryID, opts[0].ID)
	}
}

func TestDefaultStoryID(t *testing.T) {
	srv := testServer(t)
	id := srv.defaultStoryID()
	if id != testStoryID {
		t.Errorf("Expected defaultStoryID %q, got %q", testStoryID, id)
	}
}

func TestDefaultStoryID_NoStories(t *testing.T) {
	srv := &Server{Engine: &game.Engine{Stories: nil}}
	id := srv.defaultStoryID()
	if id != game.DefaultStoryID {
		t.Errorf("Expected DefaultStoryID when no stories, got %q", id)
	}
}

func TestHandleBegin_WithCookieNoFormSession(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	// No session_id in form; cookie is used
	req := httptest.NewRequest(http.MethodPost, "/begin", strings.NewReader("name=Hero&avatar=male_old&story_id=test"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	updated, ok, err := srv.Store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("Expected updated session")
	}
	if !updated.RerollUsed {
		t.Error("Expected reroll to be locked after begin")
	}
}

func TestHandleBegin_InvalidStoryIDUsesDefault(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/begin", strings.NewReader("session_id="+id+"&name=Hero&avatar=male_young&story_id=nonexistent"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
}

func TestHandlePlay_ApplyChoiceError(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer("test", "start")
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=invalid_choice"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "That choice") {
		t.Error("Expected error message about invalid choice")
	}
}

func TestHandleMap_NoSession_RedirectsToStart(t *testing.T) {
	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/map", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Errorf("GET /map no session: expected 302, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != pathStart {
		t.Errorf("GET /map no session: expected Location %s, got %q", pathStart, loc)
	}
}

func TestHandlePlay_BattleNode_ShowsEffectiveChoices(t *testing.T) {
	// Story with a node that has a battle choice so makeViewModel builds EffectiveChoices.
	story := &game.Story{
		Start: "start",
		Nodes: map[string]*game.Node{
			"start": {
				Text: "Start.",
				Choices: []game.Choice{
					{Key: "go", Text: "Go to road", Next: testNodeRoad},
				},
			},
			testNodeRoad: {
				Text: "A goblin blocks the path.",
				Choices: []game.Choice{
					{
						Key:  "fight",
						Text: "Fight",
						Battle: &game.Battle{
							Enemies: []game.Enemy{{Name: "Goblin", Strength: 8, Health: 3}},
						},
					},
				},
			},
		},
	}
	engine := &game.Engine{Stories: map[string]*game.Story{testStoryID: story}}
	store := session.NewMemoryStore[game.PlayerState]()
	tmplDir := filepath.Join("..", "..", "templates")
	tmpl := template.Must(template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "layout_head.html"),
		filepath.Join(tmplDir, "sidebar_left.html"),
		filepath.Join(tmplDir, "sidebar_right.html"),
		filepath.Join(tmplDir, "sidebar_left_oob.html"),
		filepath.Join(tmplDir, "sidebar_right_oob.html"),
		filepath.Join(tmplDir, "game.html"),
		filepath.Join(tmplDir, "game_response.html"),
		filepath.Join(tmplDir, "start.html"),
	))
	srv := &Server{Engine: engine, Store: store, Tmpl: tmpl}

	ctx := context.Background()
	st := game.NewPlayer(testStoryID, "start")
	st.NodeID = testNodeRoad // already at road
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=fight"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Attack") {
		t.Error("Expected response to contain battle choice 'Attack' from makeViewModel EffectiveChoices")
	}
}

func TestHandlePlay_BattleNode_NoRunAwayWithoutNext(t *testing.T) {
	// Battle choice without 'next' should NOT show "Run away" option.
	story := &game.Story{
		Start: "start",
		Nodes: map[string]*game.Node{
			"start": {
				Text: "Start.",
				Choices: []game.Choice{
					{Key: "go", Text: "Go to road", Next: testNodeRoad},
				},
			},
			testNodeRoad: {
				Text: "A goblin blocks the path.",
				Choices: []game.Choice{
					{
						Key:  "fight",
						Text: "Fight",
						// No Next field - running should not be offered
						Battle: &game.Battle{
							Enemies:       []game.Enemy{{Name: "Goblin", Strength: 8, Health: 3}},
							OnVictoryNext: "victory",
							OnDefeatNext:  "defeat",
						},
					},
				},
			},
			"victory": {Text: "You won!", Ending: true},
			"defeat":  {Text: "You lost!", Ending: true},
		},
	}
	engine := &game.Engine{Stories: map[string]*game.Story{testStoryID: story}}
	store := session.NewMemoryStore[game.PlayerState]()
	tmplDir := filepath.Join("..", "..", "templates")
	tmpl := template.Must(template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "layout_head.html"),
		filepath.Join(tmplDir, "sidebar_left.html"),
		filepath.Join(tmplDir, "sidebar_right.html"),
		filepath.Join(tmplDir, "sidebar_left_oob.html"),
		filepath.Join(tmplDir, "sidebar_right_oob.html"),
		filepath.Join(tmplDir, "game.html"),
		filepath.Join(tmplDir, "game_response.html"),
		filepath.Join(tmplDir, "start.html"),
	))
	srv := &Server{Engine: engine, Store: store, Tmpl: tmpl}

	ctx := context.Background()
	st := game.NewPlayer(testStoryID, "start")
	st.NodeID = testNodeRoad
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=fight"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Attack Goblin") {
		t.Error("Expected 'Attack Goblin' option")
	}
	if strings.Contains(body, "Run away") {
		t.Error("Expected NO 'Run away' option when battle choice has no 'next' field")
	}
}

func TestHandlePlay_BattleNode_RunAwayWithNext(t *testing.T) {
	// Battle choice WITH 'next' should show "Run away" option.
	story := &game.Story{
		Start: "start",
		Nodes: map[string]*game.Node{
			"start": {
				Text: "Start.",
				Choices: []game.Choice{
					{Key: "go", Text: "Go to road", Next: testNodeRoad},
				},
			},
			testNodeRoad: {
				Text: "A goblin blocks the path.",
				Choices: []game.Choice{
					{
						Key:  "fight",
						Text: "Fight",
						Next: "escaped", // Has Next - running should be offered
						Battle: &game.Battle{
							Enemies:       []game.Enemy{{Name: "Goblin", Strength: 8, Health: 3}},
							OnVictoryNext: "victory",
							OnDefeatNext:  "defeat",
						},
					},
				},
			},
			"escaped": {Text: "You escaped!", Ending: true},
			"victory": {Text: "You won!", Ending: true},
			"defeat":  {Text: "You lost!", Ending: true},
		},
	}
	engine := &game.Engine{Stories: map[string]*game.Story{testStoryID: story}}
	store := session.NewMemoryStore[game.PlayerState]()
	tmplDir := filepath.Join("..", "..", "templates")
	tmpl := template.Must(template.ParseFiles(
		filepath.Join(tmplDir, "layout.html"),
		filepath.Join(tmplDir, "layout_head.html"),
		filepath.Join(tmplDir, "sidebar_left.html"),
		filepath.Join(tmplDir, "sidebar_right.html"),
		filepath.Join(tmplDir, "sidebar_left_oob.html"),
		filepath.Join(tmplDir, "sidebar_right_oob.html"),
		filepath.Join(tmplDir, "game.html"),
		filepath.Join(tmplDir, "game_response.html"),
		filepath.Join(tmplDir, "start.html"),
	))
	srv := &Server{Engine: engine, Store: store, Tmpl: tmpl}

	ctx := context.Background()
	st := game.NewPlayer(testStoryID, "start")
	st.NodeID = testNodeRoad
	st.Stats = game.Stats{Strength: 8, Luck: 8, Health: 12}
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/play", strings.NewReader("choice=fight"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Attack Goblin") {
		t.Error("Expected 'Attack Goblin' option")
	}
	if !strings.Contains(body, "Run away") {
		t.Error("Expected 'Run away' option when battle choice has 'next' field")
	}
}

func TestHandleMap_ReturnsPDF(t *testing.T) {
	srv := testServer(t)
	ctx := context.Background()
	st := game.NewPlayer(testStoryID, "start")
	id := srv.Store.NewID()
	if err := srv.Store.Put(ctx, id, st); err != nil {
		t.Fatalf("Put: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/map", http.NoBody)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: id})
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /map: expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("GET /map: expected Content-Type application/pdf, got %q", ct)
	}
	body := rec.Body.Bytes()
	if len(body) < 8 {
		t.Errorf("GET /map: body too short")
	}
	if !strings.HasPrefix(string(body), "%PDF") {
		t.Error("GET /map: body is not a PDF (missing %PDF header)")
	}
}
