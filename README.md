# img2pdf

img2pdf is a command line tool to convert raster images to PDF.

## Installation

```bash
go install github.com/n0madic/img2pdf@latest
```

## Usage

```bash
Usage: img2pdf [options] image1 [image2] ...

Options:
  -output string
    	Specify the output file name (default "output.pdf")
  -size string
    	Specify the page size (default "A4")
```

The image filename can be specified as a glob pattern.
