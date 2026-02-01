package web

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"adventure/internal/game"
)

const sceneryTestStoryID = "test_story"

// minimalPNG returns a minimal valid 1x1 PNG (for test fixtures).
func minimalPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode minimal PNG: %v", err)
	}
	return buf.Bytes()
}

// minimalJPEG returns a minimal valid 1x1 JPEG (for test fixtures).
func minimalJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode minimal JPEG: %v", err)
	}
	return buf.Bytes()
}

func TestHandleScenery_ServesFileFromStoryDir(t *testing.T) {
	tmpDir := t.TempDir()
	sceneryDir := filepath.Join(tmpDir, sceneryTestStoryID, "scenery")
	if err := os.MkdirAll(sceneryDir, 0o750); err != nil {
		t.Fatalf("mkdir scenery: %v", err)
	}
	forestPath := filepath.Join(sceneryDir, "forest.png")
	if err := os.WriteFile(forestPath, minimalPNG(t), 0o600); err != nil {
		t.Fatalf("write forest.png: %v", err)
	}

	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{sceneryTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/scenery/"+sceneryTestStoryID+"/forest", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /scenery/%s/forest: expected 200, got %d", sceneryTestStoryID, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypePNG {
		t.Errorf("Content-Type: expected %s, got %q", contentTypePNG, ct)
	}
	body := rec.Body.Bytes()
	if len(body) < 8 {
		t.Errorf("body too short for PNG")
	}
	if !bytes.Equal(body[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		t.Errorf("body is not a valid PNG (wrong magic)")
	}
}

func TestHandleScenery_ServesJPG(t *testing.T) {
	tmpDir := t.TempDir()
	sceneryDir := filepath.Join(tmpDir, sceneryTestStoryID, "scenery")
	if err := os.MkdirAll(sceneryDir, 0o750); err != nil {
		t.Fatalf("mkdir scenery: %v", err)
	}
	// Server tries .png then .jpg; only .jpg exists so we hit the JPEG branch.
	jpgPath := filepath.Join(sceneryDir, "sunset.jpg")
	if err := os.WriteFile(jpgPath, minimalJPEG(t), 0o600); err != nil {
		t.Fatalf("write sunset.jpg: %v", err)
	}

	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{sceneryTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/scenery/"+sceneryTestStoryID+"/sunset", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /scenery/%s/sunset: expected 200, got %d", sceneryTestStoryID, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJPEG {
		t.Errorf("Content-Type: expected %s, got %q", contentTypeJPEG, ct)
	}
	if rec.Body.Len() < 2 {
		t.Errorf("body too short for JPEG")
	}
	// JPEG magic: FFD8 FF
	if !bytes.Equal(rec.Body.Bytes()[:2], []byte{0xff, 0xd8}) {
		t.Error("body is not a valid JPEG (wrong magic)")
	}
}

func TestHandleScenery_UnknownStory_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{"other": {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/scenery/unknown_story/forest", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /scenery/unknown_story/forest: expected 404, got %d", rec.Code)
	}
}

func TestHandleScenery_NonexistentFile_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	sceneryDir := filepath.Join(tmpDir, sceneryTestStoryID, "scenery")
	if err := os.MkdirAll(sceneryDir, 0o750); err != nil {
		t.Fatalf("mkdir scenery: %v", err)
	}

	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{sceneryTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodGet, "/scenery/"+sceneryTestStoryID+"/nonexistent", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /scenery/%s/nonexistent: expected 404, got %d", sceneryTestStoryID, rec.Code)
	}
}

func TestHandleScenery_PathTraversal_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{sceneryTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}

	// Call handler directly so path is not normalized by the mux (which would redirect).
	tests := []struct {
		path   string
		reason string
	}{
		{"/scenery/" + sceneryTestStoryID + "/../other/forest", "filename with .."},
		{"/scenery/" + sceneryTestStoryID + "/..", "filename .."},
		{"/scenery/" + sceneryTestStoryID + "/forest/../../../etc/passwd", "filename with path"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "http://example.com"+tt.path, http.NoBody)
		rec := httptest.NewRecorder()
		srv.handleScenery(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s (%s): expected 404, got %d", tt.path, tt.reason, rec.Code)
		}
	}
}

func TestHandleScenery_EmptyPath_NotFound(t *testing.T) {
	srv := &Server{Engine: &game.Engine{Stories: map[string]*game.Story{}}}
	req := httptest.NewRequest(http.MethodGet, "/scenery/", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /scenery/: expected 404, got %d", rec.Code)
	}
}

func TestHandleScenery_MethodNotAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	srv := &Server{
		Engine:     &game.Engine{Stories: map[string]*game.Story{sceneryTestStoryID: {Start: "a", Nodes: map[string]*game.Node{"a": {Text: "Start"}}}}},
		StoriesDir: tmpDir,
	}
	req := httptest.NewRequest(http.MethodPost, "/scenery/"+sceneryTestStoryID+"/forest", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /scenery/%s/forest: expected 405, got %d", sceneryTestStoryID, rec.Code)
	}
}

func TestHandleScenery_NilEngine_NotFound(t *testing.T) {
	srv := &Server{Engine: nil}
	req := httptest.NewRequest(http.MethodGet, "/scenery/demo/forest", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET with nil Engine: expected 404, got %d", rec.Code)
	}
}
