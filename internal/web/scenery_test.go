package web

import (
	"bytes"
	"image"
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

func TestHandleScenery_BaseDot_NotFound(t *testing.T) {
	srv := &Server{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/scenery/.", http.NoBody)
	rec := httptest.NewRecorder()
	// Call handler directly so path is not normalized by the mux (which would redirect).
	srv.handleScenery(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("handleScenery with path ending in .: expected 404, got %d", rec.Code)
	}
}

// TestGenerateScenery_forestHasSky verifies the forest generator draws sky at the top
// (not stripes only). Regression test for object-fit: cover cropping to stripes.
func TestGenerateScenery_forestHasSky(t *testing.T) {
	img, err := generateSceneryImage("forest")
	if err != nil {
		t.Fatalf("generateSceneryImage(forest): %v", err)
	}
	rgb, ok := img.(*image.RGBA)
	if !ok {
		t.Skip("generateSceneryImage did not return *image.RGBA")
		return
	}
	// pixelSky is deep purple (0x45, 0x2c, 0x5c). Top-left should be sky.
	c := rgb.RGBAAt(0, 0)
	if c.R != 0x45 || c.G != 0x2c || c.B != 0x5c {
		t.Errorf("forest top pixel (0,0): got R=%02x G=%02x B=%02x, want sky (45 2c 5c)", c.R, c.G, c.B)
	}
}
