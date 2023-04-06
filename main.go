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

var output string

func init() {
	flag.StringVar(&output, "output", "output.pdf", "Specify the output file name")
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
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
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
			fmt.Println("adding ", files[i], " ...")
			w, h, err := getImageDimensions(files[i])
			if err != nil {
				fmt.Println("Error opening file (", files[i], "): ", err)
				continue
			}
			rect := gopdf.Rect{W: float64(w), H: float64(h)}
			if rect.W > gopdf.PageSizeA4.W {
				rect.H = rect.H * gopdf.PageSizeA4.W / rect.W
				rect.W = gopdf.PageSizeA4.W
			}
			if rect.H > gopdf.PageSizeA4.H {
				rect.W = rect.W * gopdf.PageSizeA4.H / rect.H
				rect.H = gopdf.PageSizeA4.H
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
