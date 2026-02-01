// split_portraits reads a 2x2 grid image and writes four quadrant PNGs to static/avatars/.
// Usage: go run scripts/split_portraits.go <input.png>
// Output: male_young.png (top-left), female_young.png (top-right), female_old.png (bottom-left), male_old.png (bottom-right)
package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: go run scripts/split_portraits.go <input.png>\n")
		os.Exit(1)
	}
	inPath := os.Args[1]
	f, err := os.Open(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", inPath, err)
		os.Exit(1)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		os.Exit(1)
	}
	b := img.Bounds()
	// Use bounds min so we handle images with non-zero origin (e.g. after decode)
	minX, minY := b.Min.X, b.Min.Y
	w, h := b.Dx(), b.Dy()
	halfW, halfH := w/2, h/2
	// Trim bottom of bottom row to remove common white strip below the frames (~15% of image height)
	trimBottom := h * 15 / 100
	if trimBottom > halfH/2 {
		trimBottom = halfH / 2
	}
	bottomMaxY := minY + h - trimBottom
	quadrants := []image.Rectangle{
		image.Rect(minX, minY, minX+halfW, minY+halfH),                 // top-left
		image.Rect(minX+halfW, minY, minX+w, minY+halfH),               // top-right
		image.Rect(minX, minY+halfH, minX+halfW, bottomMaxY),          // bottom-left (trimmed)
		image.Rect(minX+halfW, minY+halfH, minX+w, bottomMaxY),         // bottom-right (trimmed)
	}
	// Mapping: top-left=warrior(male_young), top-right=sorceress(female_young),
	// bottom-left=wizard(male_old), bottom-right=rogue(female_old).
	// (Many 2x2 sources put wizard bottom-left and rogue bottom-right.)
	outNames := []string{"male_young.png", "female_young.png", "male_old.png", "female_old.png"}
	outDir := filepath.Join("static", "avatars")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", outDir, err)
		os.Exit(1)
	}
	for i, r := range quadrants {
		outPath := filepath.Join(outDir, outNames[i])
		if err := writeCrop(img, r, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
			os.Exit(1)
		}
		fmt.Println(outPath)
	}
}

func writeCrop(img image.Image, r image.Rectangle, path string) error {
	dx, dy := r.Dx(), r.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, dx, dy))
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			dst.Set(x, y, img.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, dst)
}
