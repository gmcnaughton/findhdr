package main

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/gmcnaughton/gofindhdr/findhdr"
)

func main() {
  inpath := "/Users/gmcnaughton/Pictures/Photos Library.photoslibrary/Masters/2017/02"
  // inpath := "./test"
  outpath := "./out"
  optlink := true

  // Create output folder
  _ = os.Mkdir(outpath, 0755)

  count := 0

  findhdr.Find(inpath, func(hdr *findhdr.Hdr) {
    for _, image := range hdr.Images() {
      count++

      link := filepath.Join(outpath, image.Info.Name())

      if optlink {
        fmt.Println("Linking", link)
        err := os.Link(image.Path, link)
        if os.IsExist(err) {
          fmt.Printf("Skipping %s (file exists)\n", link)
        } else if err != nil {
          fmt.Printf("Error linking %s\n", link)
          fmt.Println(err)
        }
      } else {
        fmt.Println(hdr)
      }
    }
    fmt.Println()
  })

  fmt.Printf("Found %d hdrs.\n", count)
}
