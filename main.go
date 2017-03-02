package main

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/gmcnaughton/gofindhdr/findhdr"
)

func main() {
  // inpath := "/Users/gmcnaughton/Pictures/Photos Library.photoslibrary/Masters/2017/02"
  inpath := "./test"
  outpath := "./out"
  optlink := false

  // Create output folder
  if optlink {
    err := os.Mkdir(outpath, 0755)
    if err != nil && !os.IsExist(err) {
      fmt.Println("Error creating out directory", err)
    }
  }

  count := 0

  findhdr.Find(inpath, func(hdr *findhdr.Hdr) {
    count++

    if optlink {
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
      fmt.Println(hdr)
    }
  })

  fmt.Printf("Found %d hdrs.\n", count)
}
