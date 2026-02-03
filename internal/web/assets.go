package web

import (
	"path/filepath"
	"strings"
)

const assetCacheControl = "public, max-age=3600"

// storyAssetCandidates validates the request path and returns possible asset paths.
func (s *Server) storyAssetCandidates(prefix, urlPath, subdir string, extensions []string) ([]string, bool) {
	if !strings.HasPrefix(urlPath, prefix) {
		return nil, false
	}

	path := strings.TrimPrefix(urlPath, prefix)
	path = strings.Trim(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, false
	}
	storyID, filename := parts[0], parts[1]

	if s.Engine == nil || s.Engine.Stories == nil || s.Engine.Stories[storyID] == nil {
		return nil, false
	}

	safeFilename := filepath.Clean(filename)
	if safeFilename == "" || safeFilename == "." || strings.Contains(safeFilename, "..") ||
		filepath.IsAbs(safeFilename) || strings.Contains(safeFilename, string(filepath.Separator)) {
		return nil, false
	}

	baseDir := filepath.Join(s.storiesBase(), storyID, subdir)
	resolved := filepath.Join(baseDir, safeFilename)
	rel, err := filepath.Rel(baseDir, resolved)
	if err != nil || strings.Contains(rel, "..") {
		return nil, false
	}

	candidates := []string{resolved}
	for _, ext := range extensions {
		candidates = append(candidates, resolved+ext)
	}

	return candidates, true
}
