package web

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// audioExtensions lists file extensions to try when the YAML value has no extension.
var audioExtensions = []string{".mp3", ".ogg", ".wav", ".m4a"}

// Content types for audio (used in handler and tests).
const (
	contentTypeMP3 = "audio/mpeg"
	contentTypeOGG = "audio/ogg"
	contentTypeWAV = "audio/wav"
	contentTypeM4A = "audio/mp4"
)

// handleAudio serves scene audio from the per-story directory
// stories/<storyID>/audio/. URL shape: /audio/<storyID>/<filename> (no extension;
// server tries .mp3, .ogg, .wav, .m4a). StoryID must be in Engine.Stories; filename
// must be safe (no path traversal).
func (s *Server) handleAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/audio/")
	path = strings.Trim(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	storyID, filename := parts[0], parts[1]

	if s.Engine == nil || s.Engine.Stories == nil || s.Engine.Stories[storyID] == nil {
		http.NotFound(w, r)
		return
	}

	safeFilename := filepath.Clean(filename)
	if safeFilename == "" || safeFilename == "." || strings.Contains(safeFilename, "..") ||
		filepath.IsAbs(safeFilename) || strings.Contains(safeFilename, string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}

	baseDir := filepath.Join(s.storiesBase(), storyID, "audio")
	resolved := filepath.Join(baseDir, safeFilename)
	rel, err := filepath.Rel(baseDir, resolved)
	if err != nil || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}

	candidates := []string{resolved}
	for _, ext := range audioExtensions {
		candidates = append(candidates, resolved+ext)
	}

	var body []byte
	var contentType string
	for _, p := range candidates {
		b, err := os.ReadFile(p) // #nosec G304 -- p is under validated baseDir (stories/<storyID>/audio)
		if err != nil {
			continue
		}
		body = b
		switch strings.ToLower(filepath.Ext(p)) {
		case ".mp3":
			contentType = contentTypeMP3
		case ".ogg":
			contentType = contentTypeOGG
		case ".wav":
			contentType = contentTypeWAV
		case ".m4a":
			contentType = contentTypeM4A
		default:
			contentType = contentTypeMP3
		}
		break
	}
	if body == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if _, err := w.Write(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
