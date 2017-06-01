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
//     findhdr -link ./out ~/Pictures/Photos\ Library.photoslibrary/Masters/2017/03
func main() {
  var inpath, outpath string

  flag.StringVar(&outpath, "link", "", "(optional) path where images should be linked")
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage: %s [options] <path to images>\n", os.Args[0])
    flag.PrintDefaults()
  }
  flag.Parse()

  // Input path can be passed in using the `-in` flag or as the first argument
  inpath = flag.Arg(0)
  if inpath == "" {
    flag.Usage()
    os.Exit(2)
  }

  // Create output folder
  if outpath != "" {
    err := os.Mkdir(outpath, 0755)
    if err != nil && !os.IsExist(err) {
      fmt.Println("Error creating out directory", err)
      os.Exit(1)
    }
  }

  count := 0

  finder := findhdr.FilePathWalker{ inpath }
  decoder := &findhdr.ExifDecoder{}
  findhdr.Find(finder, decoder, func(hdr *findhdr.Hdr) {
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

  fmt.Printf("Found %d hdrs.\n", count)
}
