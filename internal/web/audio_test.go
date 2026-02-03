package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"adventure/internal/game"
)

const audioTestStoryID = "test_story"

// minimalMP3 returns a minimal blob that looks like an MP3 (frame sync bytes) for test fixtures.
func minimalMP3(t *testing.T) []byte {
	t.Helper()
	// MPEG frame sync: 0xFF 0xFB (or 0xFF 0xFA)
	return []byte{0xff, 0xfb, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func TestHandleAudio_ServesFileFromStoryDir(t *testing.T) {
	tmpDir := t.TempDir()
	audioDir := filepath.Join(tmpDir, audioTestStoryID, "audio")
	if err := os.MkdirAll(audioDir, 0o750); err != nil {
		t.Fatalf("mkdir audio: %v", err)
	}
	ambientPath := filepath.Join(audioDir, "ambient.mp3")
	if err := os.WriteFile(ambientPath, minimalMP3(t), 0o600); err != nil {
		t.Fatalf("write ambient.mp3: %v", err)
	}

	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{audioTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/audio/"+audioTestStoryID+"/ambient", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /audio/%s/ambient: expected 200, got %d", audioTestStoryID, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypeMP3 {
		t.Errorf("Content-Type: expected %s, got %q", contentTypeMP3, ct)
	}
	if rec.Body.Len() < 4 {
		t.Errorf("body too short for MP3")
	}
}

func TestHandleAudio_UnknownStory_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{"other": {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/audio/unknown_story/ambient", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /audio/unknown_story/ambient: expected 404, got %d", rec.Code)
	}
}

func TestHandleAudio_NonexistentFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	audioDir := filepath.Join(tmpDir, audioTestStoryID, "audio")
	if err := os.MkdirAll(audioDir, 0o750); err != nil {
		t.Fatalf("mkdir audio: %v", err)
	}

	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{audioTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/audio/"+audioTestStoryID+"/nonexistent", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /audio/%s/nonexistent: expected 404, got %d", audioTestStoryID, rec.Code)
	}
}

func TestHandleAudio_PathTraversal_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{audioTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}

	tests := []struct {
		path   string
		reason string
	}{
		{"/audio/" + audioTestStoryID + "/../other/ambient", "filename with .."},
		{"/audio/" + audioTestStoryID + "/..", "filename .."},
		{"/audio/" + audioTestStoryID + "/ambient/../../../etc/passwd", "filename with path"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "http://example.com"+tt.path, http.NoBody)
		rec := httptest.NewRecorder()
		srv.handleAudio(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s (%s): expected 404, got %d", tt.path, tt.reason, rec.Code)
		}
	}
}

func TestHandleAudio_EmptyPath_NotFound(t *testing.T) {
	srv := &Server{Engine: &game.Engine{Stories: map[string]*game.Story{}}}
	req := httptest.NewRequest(http.MethodGet, "/audio/", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /audio/: expected 404, got %d", rec.Code)
	}
}

func TestHandleAudio_MethodNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{audioTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodPost, "/audio/"+audioTestStoryID+"/ambient", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /audio/%s/ambient: expected 405, got %d", audioTestStoryID, rec.Code)
	}
}

func TestHandleAudio_NilEngine_NotFound(t *testing.T) {
	srv := &Server{Engine: nil}
	req := httptest.NewRequest(http.MethodGet, "/audio/demo/ambient", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET with nil Engine: expected 404, got %d", rec.Code)
	}
}
