package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// defaultStoriesDir is the default base directory for story files (YAML and per-story scenery).
const defaultStoriesDir = "stories"

// Content types for scenery images (used in handler and tests).
const (
	contentTypePNG  = "image/png"
	contentTypeJPEG = "image/jpeg"
)

// storiesBase returns the base directory for stories (used for resolving scenery paths).
// Tests may set Server.StoriesDir to use a temp dir.
func (s *Server) storiesBase() string {
	if s.StoriesDir != "" {
		return s.StoriesDir
	}
	return defaultStoriesDir
}

// sceneryExtensions lists file extensions to try when the YAML value has no extension.
var sceneryExtensions = []string{".png", ".jpg", ".jpeg"}

// handleScenery serves scenery images from the per-story strict directory
// stories/<storyID>/scenery/. URL shape: /scenery/<storyID>/<filename> (no extension;
// server tries .png, .jpg, .jpeg). StoryID must be in Engine.Stories; filename must
// be safe (no path traversal).
func (s *Server) handleScenery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	candidates, ok := s.storyAssetCandidates("/scenery/", r.URL.Path, "scenery", sceneryExtensions)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var body []byte
	var contentType string
	for _, p := range candidates {
		b, err := os.ReadFile(p) // #nosec G304 -- p is under validated baseDir (stories/<storyID>/scenery)
		if err != nil {
			continue
		}
		body = b
		switch strings.ToLower(filepath.Ext(p)) {
		case ".jpg", ".jpeg":
			contentType = contentTypeJPEG
		default:
			contentType = contentTypePNG
		}
		break
	}
	if body == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", assetCacheControl)
	if _, err := w.Write(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
