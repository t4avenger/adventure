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
	"strings"
)

func main() {
	code := run()
	if code != 0 {
		os.Exit(code)
	}
}

func run() int {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: go run scripts/split_portraits.go <input.png>\n")
		return 1
	}
	inPath := filepath.Clean(os.Args[1])
	if strings.Contains(inPath, "..") {
		fmt.Fprintf(os.Stderr, "path must not escape current directory\n")
		return 1
	}
	f, err := os.Open(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", inPath, err)
		return 1
	}
	defer func() {
		if cErr := f.Close(); cErr != nil {
			fmt.Fprintf(os.Stderr, "close input: %v\n", cErr)
		}
	}()
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		return 1
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
		image.Rect(minX, minY, minX+halfW, minY+halfH),         // top-left
		image.Rect(minX+halfW, minY, minX+w, minY+halfH),       // top-right
		image.Rect(minX, minY+halfH, minX+halfW, bottomMaxY),   // bottom-left (trimmed)
		image.Rect(minX+halfW, minY+halfH, minX+w, bottomMaxY), // bottom-right (trimmed)
	}
	// Mapping: top-left=warrior(male_young), top-right=sorceress(female_young),
	// bottom-left=wizard(male_old), bottom-right=rogue(female_old).
	// (Many 2x2 sources put wizard bottom-left and rogue bottom-right.)
	outNames := []string{"male_young.png", "female_young.png", "male_old.png", "female_old.png"}
	outDir := filepath.Join("static", "avatars")
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", outDir, err)
		return 1
	}
	for i, r := range quadrants {
		outPath := filepath.Join(outDir, outNames[i])
		if err := writeCrop(img, r, outDir, outNames[i]); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", outPath, err)
			return 1
		}
		fmt.Println(outPath)
	}
	return 0
}

func writeCrop(img image.Image, r image.Rectangle, outDir, baseName string) (err error) {
	dx, dy := r.Dx(), r.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, dx, dy))
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			dst.Set(x, y, img.At(r.Min.X+x, r.Min.Y+y))
		}
	}
	path := filepath.Join(outDir, baseName)
	if filepath.Clean(path) != path || strings.Contains(path, "..") {
		return fmt.Errorf("invalid path")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()
	return png.Encode(f, dst)
}
