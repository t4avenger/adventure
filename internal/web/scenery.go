package web

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Valid scenery IDs (allowlist for security). Must match CSS and story YAML.
var validSceneryIDs = map[string]bool{
	"default": true, "forest": true, "river": true, "hills": true,
	"town": true, "village": true, "road": true, "shore": true,
	"bridge": true, "clearing": true, "house_inside": true,
	"castle_inside": true, "cave": true, "dungeon": true,
}

// handleScenery serves scenery images: static/scenery/{id}.png if present,
// otherwise a Go-generated ZX81-style blocky image.
func (s *Server) handleScenery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	base := path.Base(r.URL.Path)
	if base == "." || base == "/" {
		http.NotFound(w, r)
		return
	}
	id := strings.TrimSuffix(base, path.Ext(base))
	if id == "" || !validSceneryIDs[id] {
		http.NotFound(w, r)
		return
	}

	// Prefer static file if it exists (so users can drop in PNGs).
	// Build path under static/scenery and verify no path traversal (CodeQL / G304).
	baseDir := filepath.Join("static", "scenery")
	staticPath := filepath.Clean(filepath.Join(baseDir, id+".png"))
	rel, err := filepath.Rel(baseDir, staticPath)
	if err != nil || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	if b, err := os.ReadFile(staticPath); err == nil {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		if _, err := w.Write(b); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	// Generate ZX81-style blocky image.
	img, err := generateSceneryImage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// ZX Spectrum–style palette: 8×8 attribute blocks, 2 colors per block.
// Resolution 256×192 (Spectrum native), 32×24 blocks.
var (
	spectrumBlack  = color.RGBA{0, 0, 0, 255}
	spectrumGreen  = color.RGBA{0, 0xAA, 0, 255} // basic green
	spectrumBright = color.RGBA{0, 0xFF, 0, 255} // bright green
	spectrumCyan   = color.RGBA{0, 0xFF, 0xFF, 255}
	spectrumYellow = color.RGBA{0xFF, 0xFF, 0, 255}
	spectrumGrey   = color.RGBA{0x55, 0x55, 0x55, 255}
)

const blockPx = 8
const specW, specH = 256, 192 // ZX Spectrum resolution
const blocksW, blocksH = specW / blockPx, specH / blockPx

// fillBlock fills one 8×8 attribute block at block coords (bx, by) with clr.
func fillBlock(img *image.RGBA, bx, by int, clr color.RGBA) {
	for dy := 0; dy < blockPx; dy++ {
		for dx := 0; dx < blockPx; dx++ {
			x := bx*blockPx + dx
			y := by*blockPx + dy
			if x < specW && y < specH {
				img.SetRGBA(x, y, clr)
			}
		}
	}
}

// generateSceneryImage produces a ZX Spectrum / C64–style image: strict 8×8
// attribute blocks, 256×192, so each scene is clearly distinguishable.
func generateSceneryImage(id string) (image.Image, error) {
	rect := image.Rect(0, 0, specW, specH)
	img := image.NewRGBA(rect)
	for by := 0; by < blocksH; by++ {
		for bx := 0; bx < blocksW; bx++ {
			fillBlock(img, bx, by, spectrumBlack)
		}
	}

	switch id {
	case "forest":
		// Trees: vertical columns of green 8×8 blocks (Spectrum attribute columns)
		for bx := 0; bx < blocksW; bx++ {
			if bx%4 == 0 || bx%5 == 2 {
				continue
			}
			height := blocksH - 2 - (bx % 3)
			if height < 8 {
				height = 8
			}
			for by := blocksH - 1; by >= blocksH-height; by-- {
				if by >= 0 {
					fillBlock(img, bx, by, spectrumGreen)
				}
			}
		}
	case "road":
		// Horizontal path: 3 rows of grey 8×8 blocks in the middle (C64 road strip)
		for by := blocksH/2 - 1; by <= blocksH/2+1; by++ {
			if by < 0 || by >= blocksH {
				continue
			}
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, spectrumGrey)
			}
		}
	case "clearing":
		// Open patch: circle of bright green 8×8 blocks (moonlit clearing)
		cx, cy := blocksW/2, blocksH*3/4
		r := 5
		for by := 0; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				dx := bx - cx
				dy := by - cy
				if dx*dx+dy*dy <= r*r {
					fillBlock(img, bx, by, spectrumBright)
				}
			}
		}
	case "shore":
		// Two bands: sand (yellow) at bottom, water (cyan) above (Spectrum-style bands)
		for by := blocksH - 4; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, spectrumYellow)
			}
		}
		for by := blocksH - 6; by < blocksH-4; by++ {
			if by >= 0 {
				for bx := 0; bx < blocksW; bx++ {
					fillBlock(img, bx, by, spectrumCyan)
				}
			}
		}
	case "hills":
		// Layered horizon: 4 bands of green 8×8 blocks from bottom (C64/Spectrum horizon)
		for band := 0; band < 4; band++ {
			by := blocksH - 2 - band*3
			if by < 0 {
				break
			}
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, spectrumGreen)
				if by+1 < blocksH {
					fillBlock(img, bx, by+1, spectrumGreen)
				}
			}
		}
	case "bridge":
		// Single horizontal row of green blocks (bridge over black)
		by := blocksH / 2
		for bx := 0; bx < blocksW; bx++ {
			fillBlock(img, bx, by, spectrumGreen)
		}
	case "cave", "dungeon":
		// Pillars: vertical strips of grey 8×8 blocks (Spectrum 2-color blocks)
		for i := 0; i < 5; i++ {
			bx := 4 + i*6
			if bx+1 >= blocksW {
				continue
			}
			for by := 0; by < blocksH; by++ {
				fillBlock(img, bx, by, spectrumGrey)
				fillBlock(img, bx+1, by, spectrumGrey)
			}
		}
	case "house_inside", "castle_inside":
		// Room: floor = grey blocks in center, black frame (interior frame)
		for by := blocksH / 4; by < blocksH; by++ {
			for bx := 4; bx < blocksW-4; bx++ {
				fillBlock(img, bx, by, spectrumGrey)
			}
		}
		for bx := 2; bx < blocksW-2; bx++ {
			fillBlock(img, bx, blocksH/4-1, spectrumGreen)
		}
	case "town", "village":
		// Buildings: 5 rectangles of green blocks at bottom (different heights)
		buildingHeights := []int{6, 4, 8, 5, 7}
		for i, bh := range buildingHeights {
			bx := 2 + i*6
			if bx+4 >= blocksW {
				continue
			}
			for by := blocksH - bh; by < blocksH; by++ {
				if by < 0 {
					continue
				}
				for ww := 0; ww < 4 && bx+ww < blocksW; ww++ {
					fillBlock(img, bx+ww, by, spectrumGreen)
				}
			}
		}
	case "river":
		// One horizontal band of cyan (water) – very distinct
		by := blocksH / 2
		for bx := 0; bx < blocksW; bx++ {
			fillBlock(img, bx, by, spectrumCyan)
			if by+1 < blocksH {
				fillBlock(img, bx, by+1, spectrumCyan)
			}
		}
	default:
		// default: checkerboard of green blocks (Spectrum loading screen style)
		for by := 0; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				if (bx+by)%2 == 0 {
					fillBlock(img, bx, by, spectrumGreen)
				}
			}
		}
	}

	return img, nil
}
