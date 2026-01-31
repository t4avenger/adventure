package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

const contentTypePNG = "image/png"

func TestHandleScenery_GET_ReturnsPNG(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/scenery/forest.png", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET /scenery/forest.png: expected 200, got %d", rec.Code)
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

func TestHandleScenery_InvalidID_NotFound(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/scenery/nosuch.png", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /scenery/nosuch.png: expected 404, got %d", rec.Code)
	}
}

func TestHandleScenery_MethodNotAllowed(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/scenery/forest.png", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /scenery/forest.png: expected 405, got %d", rec.Code)
	}
}

func TestHandleScenery_AllValidIDs_ReturnPNG(t *testing.T) {
	ids := []string{
		"default", "forest", "river", "hills", "town", "village",
		"road", "shore", "bridge", "clearing", "house_inside",
		"castle_inside", "cave", "dungeon",
	}
	srv := &Server{}
	for _, id := range ids {
		req := httptest.NewRequest(http.MethodGet, "/scenery/"+id+".png", http.NoBody)
		rec := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET /scenery/%s.png: expected 200, got %d", id, rec.Code)
		}
		if rec.Header().Get("Content-Type") != contentTypePNG {
			t.Errorf("GET /scenery/%s.png: expected Content-Type %s", id, contentTypePNG)
		}
		if rec.Body.Len() < 100 {
			t.Errorf("GET /scenery/%s.png: body too short", id)
		}
	}
}

func TestHandleScenery_EmptyBase_NotFound(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/scenery/", http.NoBody)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /scenery/: expected 404, got %d", rec.Code)
	}
}
