package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var (
	output string
	size   string
)

func init() {
	flag.StringVar(&output, "output", "output.pdf", "Specify the output file name")
	flag.StringVar(&size, "size", "A4", "Specify the page size")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] image1 [image2] ...\n", os.Args[0])
		fmt.Fprint(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("ERROR: No source image files provided")
		flag.Usage()
		os.Exit(1)
	}

	pdf := gopdf.GoPdf{}
	pageSize := getPageSize(size)
	pdf.Start(gopdf.Config{PageSize: *pageSize})

	imagesAdded := processImages(&pdf, args, pageSize)

	if imagesAdded == 0 {
		fmt.Println("No images were successfully processed")
		os.Exit(1)
	}

	fmt.Println("saving to", output)
	if err := pdf.WritePdf(output); err != nil {
		fmt.Printf("Error writing PDF: %v\n", err)
		os.Exit(1)
	}
}

func processImages(pdf *gopdf.GoPdf, args []string, pageSize *gopdf.Rect) int {
	count := 0
	for _, arg := range args {
		files, err := getFiles(arg)
		if err != nil {
			fmt.Printf("Error with file mask (%s): %v\n", arg, err)
			continue
		}

		for _, file := range files {
			fmt.Printf("adding %s...\n", file)
			if err := addImageToPDF(pdf, file, pageSize); err != nil {
				fmt.Printf("Error processing file (%s): %v\n", file, err)
			} else {
				count++
			}
		}
	}
	return count
}

func getFiles(arg string) ([]string, error) {
	if strings.Contains(arg, "*") {
		return filepath.Glob(arg)
	}
	return []string{arg}, nil
}

func addImageToPDF(pdf *gopdf.GoPdf, file string, pageSize *gopdf.Rect) error {
	w, h, err := getImageDimensions(file)
	if err != nil {
		return err
	}

	// Convert TIFF/WebP to PNG for gopdf compatibility
	rect := fitImageToPage(float64(w), float64(h), pageSize)
	pdf.AddPage()

	ext := strings.ToLower(filepath.Ext(file))
	if ext == ".tiff" || ext == ".tif" || ext == ".webp" {
		// Convert to PNG in memory
		pngData, err := convertToPNG(file)
		if err != nil {
			return fmt.Errorf("failed to convert image: %w", err)
		}
		// Use ImageFrom with in-memory data
		holder, err := gopdf.ImageHolderByReader(bytes.NewReader(pngData))
		if err != nil {
			return fmt.Errorf("failed to create image holder: %w", err)
		}
		return pdf.ImageByHolder(holder, 0, 0, &rect)
	}

	// Use regular Image method for already supported formats
	return pdf.Image(file, 0, 0, &rect)
}

func fitImageToPage(w, h float64, pageSize *gopdf.Rect) gopdf.Rect {
	rect := gopdf.Rect{W: w, H: h}
	if rect.W > pageSize.W {
		rect.H = rect.H * pageSize.W / rect.W
		rect.W = pageSize.W
	}
	if rect.H > pageSize.H {
		rect.W = rect.W * pageSize.H / rect.H
		rect.H = pageSize.H
	}
	return rect
}

func getPageSize(s string) *gopdf.Rect {
	sizes := map[string]*gopdf.Rect{
		"a0":        gopdf.PageSizeA0,
		"a1":        gopdf.PageSizeA1,
		"a2":        gopdf.PageSizeA2,
		"a3":        gopdf.PageSizeA3,
		"a4":        gopdf.PageSizeA4,
		"a4l":       gopdf.PageSizeA4Landscape,
		"a4s":       gopdf.PageSizeA4Small,
		"a5":        gopdf.PageSizeA5,
		"b4":        gopdf.PageSizeB4,
		"b5":        gopdf.PageSizeB5,
		"executive": gopdf.PageSizeExecutive,
		"folio":     gopdf.PageSizeFolio,
		"legal":     gopdf.PageSizeLegal,
		"ledger":    gopdf.PageSizeLedger,
		"letter":    gopdf.PageSizeLetter,
		"quarto":    gopdf.PageSizeQuarto,
		"statement": gopdf.PageSizeStatement,
		"tabloid":   gopdf.PageSizeTabloid,
		"10x14":     gopdf.PageSize10x14,
	}

	if size, ok := sizes[strings.ToLower(s)]; ok {
		return size
	}
	return gopdf.PageSizeA4
}

func getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	// DecodeConfig is used to get image dimensions without fully decoding the image
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}

func convertToPNG(filePath string) ([]byte, error) {
	// Open source image
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Convert to 8-bit RGBA to handle 16-bit images
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Encode to PNG in memory
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
