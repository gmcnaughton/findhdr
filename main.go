package main

import (
  "fmt"
  "os"
  "path/filepath"

  "github.com/gmcnaughton/gofindhdr/findhdr"
)

func main() {
  inpath := "/Users/gmcnaughton/Pictures/Photos Library.photoslibrary/Masters/2017/02/27"
  // inpath := "./test"
  outpath := "./out"

  // Create output folder
  _ = os.Mkdir(outpath, 0755)

  findhdr.Find(inpath, func(hdr *findhdr.Hdr) {
    fmt.Println(hdr)
    fmt.Println("-----------")
    for _, image := range hdr.Images() {
      link := filepath.Join(outpath, image.Info.Name())
      fmt.Println("  Linking", link)

      err := os.Link(image.Path, link)
      if os.IsExist(err) {
        fmt.Printf("  Skipping %s (file already exists)\n", link)
      } else if err != nil {
        fmt.Printf("  Error linking %s\n", link)
        fmt.Println(err)
      }
    }
    fmt.Println()
  })
}
