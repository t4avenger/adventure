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

// Pixel-art sunset palette: warm/cool tones to match static harbor scenery.
// Resolution 256×192, 8×8 blocks (blocky pixel-art style).
var (
	pixelBlack  = color.RGBA{0x18, 0x14, 0x28, 255} // dark purple-black
	pixelSky    = color.RGBA{0x45, 0x2c, 0x5c, 255} // deep purple sky
	pixelWater  = color.RGBA{0x2d, 0x3a, 0x5c, 255} // deep blue
	pixelSand   = color.RGBA{0x8b, 0x73, 0x55, 255} // warm tan
	pixelStone  = color.RGBA{0x55, 0x55, 0x66, 255} // grey stone
	pixelGreen  = color.RGBA{0x2d, 0x5a, 0x3d, 255} // muted green (trees)
	pixelBright = color.RGBA{0x6b, 0x8c, 0x5a, 255} // lighter green (clearing)
	pixelWarm   = color.RGBA{0xc4, 0x6c, 0x32, 255} // warm brown/orange (windows, path)
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

// generateSceneryImage produces a pixel-art style image (8×8 blocks, 256×192)
// with a sunset/harbor-style palette so generated scenes match static shore/town assets.
func generateSceneryImage(id string) (image.Image, error) {
	rect := image.Rect(0, 0, specW, specH)
	img := image.NewRGBA(rect)
	for by := 0; by < blocksH; by++ {
		for bx := 0; bx < blocksW; bx++ {
			fillBlock(img, bx, by, pixelBlack)
		}
	}

	switch id {
	case "forest":
		// Illustrative forest: sky, dark-forest mid, ground, and clear tree shapes (canopy + trunk).
		// Sky band (twilight purple) – unmistakably different from old stripe-only image.
		for by := 0; by < blocksH/3; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		// Middle: dark forest background (not black stripes – reads as depth).
		for by := blocksH / 3; by < blocksH-2; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelBlack)
			}
		}
		// Forest floor (bottom two rows, green).
		for by := blocksH - 2; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
			}
		}
		// Trees: trunk (1 block) + canopy (3 blocks wide, 4 rows). Positions avoid uniform stripes.
		treeCols := []int{2, 9, 16, 23, 6, 19}
		for _, bx := range treeCols {
			if bx < 1 || bx >= blocksW-1 {
				continue
			}
			for by := blocksH - 3; by < blocksH; by++ {
				fillBlock(img, bx, by, pixelStone) // trunk: grey-brown so it reads as bark
			}
			canopyTop := blocksH - 7
			for _, dx := range []int{-1, 0, 1} {
				cx := bx + dx
				if cx < 0 || cx >= blocksW {
					continue
				}
				for by := canopyTop; by < blocksH-3; by++ {
					fillBlock(img, cx, by, pixelGreen)
				}
			}
			if canopyTop >= 0 {
				fillBlock(img, bx+1, canopyTop, pixelBright)
			}
		}
	case "road":
		// Road through landscape: sky, green sides, stone path (illustrative).
		for by := 0; by < blocksH/3; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		for by := blocksH / 3; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
			}
		}
		// Cobbled road strip (stone) down the middle.
		for by := blocksH/2 - 1; by <= blocksH/2+1; by++ {
			if by < 0 || by >= blocksH {
				continue
			}
			for bx := 4; bx < blocksW-4; bx++ {
				fillBlock(img, bx, by, pixelStone)
			}
		}
	case "clearing":
		// Open patch: sky, ring of trees, circular clearing (lighter green) with ground.
		for by := 0; by < blocksH/4; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		for by := blocksH / 4; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
			}
		}
		// Clearing circle (lighter green).
		cx, cy := blocksW/2, blocksH*3/4
		r := 5
		for by := 0; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				dx := bx - cx
				dy := by - cy
				if dx*dx+dy*dy <= r*r {
					fillBlock(img, bx, by, pixelBright)
				}
			}
		}
		// A few small tree silhouettes at edges (so it reads as clearing in woods).
		for _, bx := range []int{2, 26} {
			for by := blocksH - 5; by < blocksH; by++ {
				fillBlock(img, bx, by, pixelGreen)
			}
			if bx+1 < blocksW {
				for by := blocksH - 4; by < blocksH-1; by++ {
					fillBlock(img, bx+1, by, pixelGreen)
				}
			}
		}
	case "shore":
		// Fallback if static shore.png missing: sand + water bands (sunset palette)
		for by := blocksH - 4; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSand)
			}
		}
		for by := blocksH - 6; by < blocksH-4; by++ {
			if by >= 0 {
				for bx := 0; bx < blocksW; bx++ {
					fillBlock(img, bx, by, pixelWater)
				}
			}
		}
	case "hills":
		// Hills at dusk: sky, then layered green hills (illustrative depth).
		for by := 0; by < blocksH/4; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		for band := 0; band < 4; band++ {
			by := blocksH - 2 - band*4
			if by < blocksH/4 {
				break
			}
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
				if by+1 < blocksH {
					fillBlock(img, bx, by+1, pixelGreen)
				}
			}
		}
	case "bridge":
		// Stone bridge strip over dark
		by := blocksH / 2
		for bx := 0; bx < blocksW; bx++ {
			fillBlock(img, bx, by, pixelStone)
		}
	case "cave", "dungeon":
		// Pillars: vertical strips of stone (dungeon)
		for i := 0; i < 5; i++ {
			bx := 4 + i*6
			if bx+1 >= blocksW {
				continue
			}
			for by := 0; by < blocksH; by++ {
				fillBlock(img, bx, by, pixelStone)
				fillBlock(img, bx+1, by, pixelStone)
			}
		}
	case "house_inside", "castle_inside":
		// Room: stone floor, warm strip (window light)
		for by := blocksH / 4; by < blocksH; by++ {
			for bx := 4; bx < blocksW-4; bx++ {
				fillBlock(img, bx, by, pixelStone)
			}
		}
		for bx := 2; bx < blocksW-2; bx++ {
			fillBlock(img, bx, blocksH/4-1, pixelWarm)
		}
	case "town", "village":
		// Fallback if static town.png missing: building blocks (stone + warm windows)
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
					fillBlock(img, bx+ww, by, pixelStone)
				}
			}
			// Warm window strip
			by := blocksH - bh - 1
			if by >= 0 {
				for ww := 0; ww < 4 && bx+ww < blocksW; ww++ {
					fillBlock(img, bx+ww, by, pixelWarm)
				}
			}
		}
	case "river":
		// River through landscape: sky, green banks, horizontal water band (illustrative).
		for by := 0; by < blocksH/3; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		for by := blocksH / 3; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
			}
		}
		// Water band (deep blue).
		for by := blocksH/2 - 1; by <= blocksH/2+2; by++ {
			if by < 0 || by >= blocksH {
				continue
			}
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelWater)
			}
		}
	default:
		// default: sky + ground bands (sunset mood)
		for by := 0; by < blocksH/2; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelSky)
			}
		}
		for by := blocksH / 2; by < blocksH; by++ {
			for bx := 0; bx < blocksW; bx++ {
				fillBlock(img, bx, by, pixelGreen)
			}
		}
	}

	return img, nil
}
