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
//     findhdr -in ./test
//     findhdr -in ~/Pictures/Photos\ Library.photoslibrary/Masters/2017/03 -out ./out -link
func main() {
  var inpath, outpath string
  var link bool

  // flag.StringVar(&inpath, "in", "/Users/gmcnaughton/Pictures/Photos Library.photoslibrary/Masters/2017/02", "path to input directory to search")
  flag.StringVar(&inpath, "in", "./test", "path to search")
  flag.StringVar(&outpath, "out", "./out", "path where matches should be linked")
  flag.BoolVar(&link, "link", false, "true if matches should be linked")
  flag.Parse()

  // Create output folder
  if link {
    err := os.Mkdir(outpath, 0755)
    if err != nil && !os.IsExist(err) {
      fmt.Println("Error creating out directory", err)
    }
  }

  count := 0

  finder := findhdr.FilePathWalker{ inpath }
  decoder := &findhdr.ExifDecoder{}
  findhdr.Find(finder, decoder, func(hdr *findhdr.Hdr) {
    count++

    if link {
      for _, image := range hdr.Images() {
        link := filepath.Join(outpath, image.Info.Name())
        fmt.Println("Linking", link)
        err := os.Link(image.Path, link)
        if os.IsExist(err) {
          fmt.Println("Skipping", err)
        } else if err != nil {
          fmt.Println("Error linking", err)
        }
      }
    } else {
      fmt.Println("Found", hdr)
    }
  })

  fmt.Printf("Found %d hdrs.\n", count)
}
