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

	candidates, ok := s.storyAssetCandidates("/audio/", r.URL.Path, "audio", audioExtensions)
	if !ok {
		http.NotFound(w, r)
		return
	}

	var file *os.File
	var fileInfo os.FileInfo
	var filePath string
	var contentType string
	for _, p := range candidates {
		f, err := os.Open(p) // #nosec G304 -- p is under validated baseDir (stories/<storyID>/audio)
		if err != nil {
			continue
		}
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			_ = f.Close()
			continue
		}
		file = f
		fileInfo = info
		filePath = p
		contentType = contentTypeMP3
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
	if file == nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", assetCacheControl)
	http.ServeContent(w, r, filepath.Base(filePath), fileInfo.ModTime(), file)
}
