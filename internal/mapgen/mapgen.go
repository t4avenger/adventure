// Package mapgen builds a path of visited locations and generates a printable
// PDF treasure map (old-map style) with illustrated scenes per place.
package mapgen

import (
	"bytes"
	"math"
	"strings"

	"adventure/internal/game"

	"github.com/jung-kurt/gofpdf/v2"
)

const (
	pageW     = 595
	pageH     = 842
	margin    = 40
	sceneSize = 56.0
	pathStep  = 70.0
	fontSize  = 8
	titleSize = 16
	labelSize = 7
)

// Generate returns PDF bytes for a treasure map: visited nodes as illustrated
// scenes (beach, forest, bridge, battle, etc.) along a path. If visitedNodes
// is nil or empty, currentID is used as the only stop.
func Generate(st *game.Story, visitedNodes []string, currentID, title string) ([]byte, error) {
	if st == nil || st.Nodes == nil {
		return nil, nil
	}
	path := visitedNodes
	if len(path) == 0 {
		path = []string{currentID}
	}
	// Build list of stops with scenery and battle flag
	type stop struct {
		id       string
		scenery  string
		isBattle bool
	}
	stops := make([]stop, 0, len(path))
	for _, id := range path {
		n := st.Nodes[id]
		scenery := "default"
		isBattle := false
		if n != nil {
			if n.Scenery != "" {
				scenery = n.Scenery
			}
			for i := range n.Choices {
				if n.Choices[i].Battle != nil {
					isBattle = true
					break
				}
			}
		}
		stops = append(stops, stop{id: id, scenery: scenery, isBattle: isBattle})
	}
	// Layout: winding path (snake) so the journey zig-zags across the map
	positions := make([][2]float64, len(stops))
	x0 := float64(margin) + sceneSize
	y0 := float64(margin) + 72
	perRow := 4
	for i := range stops {
		row := i / perRow
		col := i % perRow
		if row%2 == 1 {
			col = perRow - 1 - col
		}
		positions[i][0] = x0 + float64(col)*pathStep
		positions[i][1] = y0 + float64(row)*pathStep
	}

	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// Parchment background
	pdf.SetFillColor(245, 235, 210)
	pdf.Rect(0, 0, pageW, pageH, "F")

	// Wavy / tattered black border (organic treasure-map edge)
	drawWavyBorder(pdf)

	// Brown ink for text and accents
	pdf.SetDrawColor(80, 50, 30)
	pdf.SetTextColor(80, 50, 30)
	pdf.SetLineWidth(1)

	// Title "Treasure Map" upper right, decorative
	pdf.SetFont("Helvetica", "B", titleSize)
	pdf.SetXY(pageW-margin-140, margin+2)
	pdf.CellFormat(140, 14, "Treasure Map", "", 0, "R", false, 0, "")
	if title != "" {
		pdf.SetFont("Helvetica", "", fontSize)
		pdf.SetXY(pageW-margin-140, margin+18)
		pdf.CellFormat(140, 10, title, "", 0, "R", false, 0, "")
	}

	// Compass rose (upper right, below title)
	drawCompassRose(pdf, pageW-margin-55, margin+50)

	// Dashed red path connecting scenes (winding journey)
	pdf.SetDrawColor(180, 40, 40)
	pdf.SetLineWidth(2)
	pdf.SetDashPattern([]float64{10, 6}, 0)
	for i := 0; i < len(positions)-1; i++ {
		x1, y1 := positions[i][0], positions[i][1]
		x2, y2 := positions[i+1][0], positions[i+1][1]
		pdf.Line(x1, y1, x2, y2)
	}
	pdf.SetDashPattern([]float64{}, 0)
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(80, 50, 30)

	// Illustrated scenes with labels below (cartoony, bold outlines)
	for i := range stops {
		x, y := positions[i][0], positions[i][1]
		isCurrent := stops[i].id == currentID
		drawScene(pdf, x, y, stops[i].scenery, stops[i].isBattle, isCurrent)
		// Label below scene: humanized node ID in caps (e.g. "SKULL ROCK")
		label := strings.ReplaceAll(stops[i].id, "_", " ")
		label = strings.ToUpper(label)
		if len(label) > 18 {
			label = label[:15] + "..."
		}
		pdf.SetFont("Helvetica", "B", labelSize)
		pdf.SetTextColor(40, 25, 15)
		pdf.SetXY(x-sceneSize/2-4, y+sceneSize/2+4)
		pdf.CellFormat(sceneSize+8, 10, label, "", 0, "C", false, 0, "")
		if isCurrent {
			pdf.SetFont("Helvetica", "I", 7)
			pdf.SetXY(x-sceneSize/2, y+sceneSize/2+14)
			pdf.CellFormat(sceneSize, 8, "You are here", "", 0, "C", false, 0, "")
		}
		pdf.SetFont("Helvetica", "", fontSize)
		pdf.SetTextColor(80, 50, 30)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawWavyBorder draws an organic, tattered black border around the map (parchment edge).
func drawWavyBorder(pdf *gofpdf.Fpdf) {
	pts := wavyRectPoints(margin, margin, pageW-2*margin, pageH-2*margin, 12, 4)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(2)
	pdf.Polygon(pts, "D")
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(80, 50, 30)
}

// wavyRectPoints returns polygon points for a rectangle with sinusoidal wobble on each side.
func wavyRectPoints(x, y, w, h float64, steps int, amp float64) []gofpdf.PointType {
	pts := make([]gofpdf.PointType, 0, steps*4+4)
	// Top edge (left to right)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pts = append(pts, gofpdf.PointType{
			X: x + t*w + amp*math.Sin(float64(i)*0.7),
			Y: y + amp*math.Cos(float64(i)*0.5),
		})
	}
	// Right edge (top to bottom)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pts = append(pts, gofpdf.PointType{
			X: x + w + amp*math.Sin(float64(i)*0.6),
			Y: y + t*h + amp*math.Cos(float64(i)*0.4),
		})
	}
	// Bottom edge (right to left)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pts = append(pts, gofpdf.PointType{
			X: x + w - t*w + amp*math.Sin(float64(i)*0.8),
			Y: y + h + amp*math.Cos(float64(i)*0.3),
		})
	}
	// Left edge (bottom to top), ending at (x,y) so polygon closes
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		pts = append(pts, gofpdf.PointType{
			X: x + amp*math.Sin(float64(i)*0.5),
			Y: y + h - t*h + amp*math.Cos(float64(i)*0.6),
		})
	}
	return pts
}

// drawCompassRose draws an eight-point compass rose with N/S/E/W labels (red/yellow/brown).
func drawCompassRose(pdf *gofpdf.Fpdf, cx, cy float64) {
	const rad = 22.0
	// Outer circle (brown)
	pdf.SetDrawColor(101, 67, 33)
	pdf.SetLineWidth(1)
	pdf.Circle(cx, cy, rad, "D")
	// Eight points: N, NE, E, SE, S, SW, W, NW
	for i := 0; i < 8; i++ {
		angle := float64(i)*45.0*math.Pi/180 - math.Pi/2 // 0 = N
		dx := rad * math.Cos(angle)
		dy := rad * math.Sin(angle)
		if i%2 == 0 {
			pdf.SetDrawColor(180, 40, 40) // red for cardinal
			pdf.SetLineWidth(1.5)
		} else {
			pdf.SetDrawColor(180, 140, 60) // yellow/brown for ordinal
			pdf.SetLineWidth(1)
		}
		pdf.Line(cx, cy, cx+dx, cy+dy)
	}
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(80, 50, 30)
	// Labels N, S, E, W
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(80, 50, 30)
	for _, lab := range []struct {
		label  string
		dx, dy float64
	}{
		{"N", 0, -rad - 10},
		{"S", 0, rad + 10},
		{"E", rad + 8, 0},
		{"W", -rad - 8, 0},
	} {
		pdf.SetXY(cx+lab.dx-4, cy+lab.dy-3)
		pdf.CellFormat(8, 6, lab.label, "", 0, "C", false, 0, "")
	}
	pdf.SetFont("Helvetica", "", fontSize)
}

// drawScene draws a small pictorial at (x,y) for the given scenery and battle flag (bold black outlines).
func drawScene(pdf *gofpdf.Fpdf, x, y float64, scenery string, isBattle, isCurrent bool) {
	r := sceneSize / 2.0
	if isCurrent {
		pdf.SetDrawColor(80, 50, 20)
		pdf.SetLineWidth(2)
		pdf.Circle(x, y, r+4.0, "D")
		pdf.SetLineWidth(1)
	}
	// Cartoony hand-drawn look: black outlines for the scene
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(1.2)
	switch scenery {
	case "shore":
		drawShore(pdf, x, y, r)
	case "forest":
		drawForest(pdf, x, y, r)
	case "road":
		drawRoad(pdf, x, y, r)
	case "bridge":
		drawBridge(pdf, x, y, r)
	case "clearing":
		drawClearing(pdf, x, y, r)
	case "cave", "dungeon":
		drawCave(pdf, x, y, r)
	case "river":
		drawRiver(pdf, x, y, r)
	case "hills":
		drawHills(pdf, x, y, r)
	case "town", "village":
		drawTown(pdf, x, y, r)
	case "house_inside", "castle_inside":
		drawHouse(pdf, x, y, r)
	default:
		drawDefault(pdf, x, y, r)
	}
	pdf.SetLineWidth(1)
	pdf.SetDrawColor(80, 50, 30)
	if isBattle {
		drawBattle(pdf, x, y, r)
	}
}

func drawShore(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Waves and sand: wavy line, then sun
	for i := 0; i < 5; i++ {
		dx := -r + float64(i)*r*0.5
		dy := 3 * float64(i%2)
		pdf.Line(x+dx, y+dy, x+dx+r*0.5, y-dy)
	}
	pdf.Circle(x+r*0.3, y-r*0.4, 4, "D")
}

func drawForest(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Trees: vertical lines with small circles on top
	for i, dx := range []float64{-r * 0.4, 0, r * 0.35} {
		h := 12 + float64(i)*4
		pdf.Line(x+dx, y, x+dx, y-h)
		pdf.Circle(x+dx, y-h, 5, "D")
	}
}

func drawRoad(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Horizontal path band
	pdf.SetLineWidth(2)
	pdf.Line(x-r*0.8, y, x+r*0.8, y)
	pdf.SetLineWidth(1)
}

func drawBridge(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Arc over gap (bridge)
	pdf.Line(x-r*0.7, y+3, x+r*0.7, y+3)
	pdf.Arc(x, y+8, 20, 8, 0, 0, 180, "D")
}

func drawClearing(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.Circle(x, y, r*0.5, "D")
	pdf.Circle(x, y, r*0.25, "D")
}

func drawCave(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.Arc(x, y+r*0.3, r*0.8, r*0.6, 0, 0, 180, "D")
	pdf.Line(x-r*0.8, y+r*0.3, x-r*0.8, y+r)
	pdf.Line(x+r*0.8, y+r*0.3, x+r*0.8, y+r)
}

func drawRiver(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.SetLineWidth(2)
	pdf.Line(x-r, y, x+r, y)
	for i := -1; i <= 1; i++ {
		dx := float64(i) * r * 0.4
		pdf.Line(x+dx, y-4, x+dx+8, y+4)
	}
	pdf.SetLineWidth(1)
}

func drawHills(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.Arc(x-r*0.5, y+r*0.2, r*0.6, r*0.4, 0, 0, 180, "D")
	pdf.Arc(x, y+r*0.3, r*0.5, r*0.35, 0, 0, 180, "D")
	pdf.Arc(x+r*0.4, y+r*0.25, r*0.5, r*0.35, 0, 0, 180, "D")
}

func drawTown(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Small rectangles (buildings)
	for i, dx := range []float64{-r * 0.5, -r * 0.1, r * 0.3} {
		w, h := 10.0, 14.0+float64(i)*4
		pdf.Rect(x+dx-w/2, y+h-r*0.3, w, -h, "D")
	}
}

func drawHouse(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.Rect(x-r*0.4, y-r*0.2, r*0.8, r*0.6, "D")
	pdf.Line(x-r*0.4, y-r*0.2, x, y-r*0.5)
	pdf.Line(x, y-r*0.5, x+r*0.4, y-r*0.2)
}

func drawDefault(pdf *gofpdf.Fpdf, x, y, r float64) {
	pdf.Circle(x, y, r*0.35, "D")
}

func drawBattle(pdf *gofpdf.Fpdf, x, y, r float64) {
	// Crossed lines (swords) over the scene
	pdf.SetLineWidth(1.5)
	pdf.Line(x-r*0.4, y-r*0.4, x+r*0.4, y+r*0.4)
	pdf.Line(x-r*0.4, y+r*0.4, x+r*0.4, y-r*0.4)
	pdf.SetLineWidth(1)
}
