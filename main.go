package main

import (
  "github.com/gmcnaughton/gofindhdr/findhdr"
)

func main() {
  var path string
  if (len(os.Args) > 1) {
    path = os.Args[1]
  } else {
    path = "/Users/gmcnaughton/Pictures/Photos Library.photoslibrary/Masters/2017/02/27"
  }

  findhdr.Find(path)
}
