package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/signintech/gopdf"
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
	var err error
	for _, arg := range args {
		files := []string{arg}
		if strings.Contains(arg, "*") {
			files, err = filepath.Glob(arg)
			if err != nil {
				fmt.Println("Error with file mask (", arg, "): ", err)
				continue
			}
		}
		for i := 0; i < len(files); i++ {
			fmt.Println("adding", files[i], "...")
			w, h, err := getImageDimensions(files[i])
			if err != nil {
				fmt.Println("Error opening file (", files[i], "): ", err)
				continue
			}
			rect := gopdf.Rect{W: float64(w), H: float64(h)}
			if rect.W > pageSize.W {
				rect.H = rect.H * pageSize.W / rect.W
				rect.W = pageSize.W
			}
			if rect.H > pageSize.H {
				rect.W = rect.W * pageSize.H / rect.H
				rect.H = pageSize.H
			}
			pdf.AddPage()
			pdf.Image(files[i], 0, 0, &rect)
		}
	}
	if pdf.GetNumberOfPages() == 0 {
		fmt.Println("No images found")
		os.Exit(1)
	}
	fmt.Println("saving to", output)
	pdf.WritePdf(output)
}

func getPageSize(s string) *gopdf.Rect {
	// Detect page size
	switch strings.ToLower(s) {
	case "a0":
		return gopdf.PageSizeA0
	case "a1":
		return gopdf.PageSizeA1
	case "a2":
		return gopdf.PageSizeA2
	case "a3":
		return gopdf.PageSizeA3
	case "a4":
		return gopdf.PageSizeA4
	case "a4l":
		return gopdf.PageSizeA4Landscape
	case "a4s":
		return gopdf.PageSizeA4Small
	case "a5":
		return gopdf.PageSizeA5
	case "b4":
		return gopdf.PageSizeB4
	case "b5":
		return gopdf.PageSizeB5
	case "executive":
		return gopdf.PageSizeExecutive
	case "folio":
		return gopdf.PageSizeFolio
	case "legal":
		return gopdf.PageSizeLegal
	case "ledger":
		return gopdf.PageSizeLedger
	case "letter":
		return gopdf.PageSizeLetter
	case "quarto":
		return gopdf.PageSizeQuarto
	case "statement":
		return gopdf.PageSizeStatement
	case "tabloid":
		return gopdf.PageSizeTabloid
	case "10x14":
		return gopdf.PageSize10x14
	}
	return gopdf.PageSizeA4
}

func getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}
