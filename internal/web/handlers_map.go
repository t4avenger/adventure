package web

import (
	"net/http"

	"adventure/internal/mapgen"
)

func (s *Server) handleMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	id := s.sessionID(r)
	if id == "" {
		http.Redirect(w, r, "/start", http.StatusFound)
		return
	}
	state, ok, err := s.Store.Get(ctx, id)
	if err != nil || !ok {
		http.Redirect(w, r, "/start", http.StatusFound)
		return
	}
	st := s.Engine.Stories[state.StoryID]
	if st == nil {
		http.Redirect(w, r, "/start", http.StatusFound)
		return
	}
	title := st.Title
	if title == "" {
		title = state.StoryID
	}
	storiesDir := s.storiesBase()
	pdf, err := mapgen.Generate(st, state.VisitedNodes, state.NodeID, title, state.StoryID, storiesDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="adventure-map.pdf"`)
	if _, err := w.Write(pdf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
