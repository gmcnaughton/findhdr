package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gmcnaughton/findhdr"
)

// Usage:
//     go build ...findhdr
//     go install ...findhdr
//     findhdr ./test
//     findhdr -min 3 -max 5 ./test
//     findhdr -link ./out ~/Pictures/Photos\ Library.photoslibrary/Masters/2017/03
func main() {
	var inpath, outpath string
	var min, max int

	flag.StringVar(&outpath, "link", "", "(optional) path where images should be linked")
	flag.IntVar(&min, "min", 3, "(optional) min number of images to consider")
	flag.IntVar(&max, "max", 3, "(optional) max number of images to consider")
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "Usage: %s [options] <path to images>\n", os.Args[0])
		if err != nil {
			return
		}
		flag.PrintDefaults()
	}
	flag.Parse()

	// Input path can be passed in using the `-in` flag or as the first argument
	inpath = flag.Arg(0)
	if inpath == "" || min < 1 || max < 1 || min > max {
		flag.Usage()
		os.Exit(2)
	}

	// Create output folder
	if outpath != "" {
		err := os.MkdirAll(outpath, 0700)
		if err != nil && !os.IsExist(err) {
			fmt.Println("Error creating out directory", err)
			os.Exit(1)
		}
	}

	count := 0

	finder := findhdr.NewFileFinder(inpath)
	decoder := findhdr.NewDecoder()
	err := findhdr.Find(finder, decoder, min, max, func(hdr *findhdr.Hdr) {
		count++

		if outpath != "" {
			for _, image := range hdr.Images() {
				link := filepath.Join(outpath, image.Info.Name())
				err := os.Link(image.Path, link)
				if os.IsExist(err) {
					fmt.Println("Skipping", link, "file exists")
				} else if err != nil {
					fmt.Println("Error linking", err)
					os.Exit(1)
				} else {
					fmt.Println("Linking", link)
				}
			}
		} else {
			fmt.Println("Found", hdr)
		}
	})
	if err != nil {
		fmt.Println("Error trying to search directory", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d hdrs.\n", count)
}
