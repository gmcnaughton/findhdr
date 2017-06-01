package findhdr

import (
  "testing"
  "os"
)

// Usage:
//     go test -v ./...

func TestHdrIsHdrWhenEmpty(t *testing.T) {
  hdr := Hdr{}
  if hdr.IsHdr() {
    t.Error("Expected empty Hdr not to be IsHdr()")
  }
}

type fixtureFile struct {
  path string
  info os.FileInfo
  err error
}

type fixtureFileFinder struct {
  files []fixtureFile
  err error
}

func (f fixtureFileFinder) Find(fileFinderFn FileFinderFunc) {
  if f.err != nil {
    fileFinderFn("", nil, f.err)
  } else {
    for _, f := range f.files {
      fileFinderFn(f.path, f.info, f.err)
    }
  }
}

func TestFindNonExistantDirectory(t *testing.T) {
  Find(fixtureFileFinder{ err: os.ErrNotExist }, &ExifDecoder{}, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindEmptyDirectory(t *testing.T) {
  Find(fixtureFileFinder{ }, &ExifDecoder{}, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindNonImageFiles(t *testing.T) {
  files := []fixtureFile{
    fixtureFile{ path: "foo1.txt" },
    fixtureFile{ path: "foo2.txt" },
    fixtureFile{ path: "foo3.txt" },
  }

  Find(fixtureFileFinder{ files: files }, &ExifDecoder{}, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindNonExistantImageFiles(t *testing.T) {
  files := []fixtureFile{
    fixtureFile{ path: "foo1.JPG" },
    fixtureFile{ path: "foo2.JPG" },
    fixtureFile{ path: "foo3.JPG" },
  }

  Find(fixtureFileFinder{ files: files }, &ExifDecoder{}, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

type fixtureDecoder struct {
  exifs []Exif
  errs []error
  calls int
}

func (decoder *fixtureDecoder) Decode(path string) (exif Exif, err error) {
  decoder.calls++
  return decoder.exifs[decoder.calls-1], decoder.errs[decoder.calls-1]
}

type fixtureExif struct {
  xdim int
  ydim int
  bias string
}

func (exif *fixtureExif) PixelXDimension() (val int, err error) {
  return exif.xdim, nil
}

func (exif *fixtureExif) PixelYDimension() (val int, err error) {
  return exif.xdim, nil
}

func (exif *fixtureExif) ExposureBiasValue() (val string, err error) {
  return exif.bias, nil
}

func TestFindSuccess(t *testing.T) {
  files := []fixtureFile{
    fixtureFile{ path: "foo1.JPG" },
    fixtureFile{ path: "foo2.JPG" },
    fixtureFile{ path: "foo3.JPG" },
  }

  exifs := []Exif {
    &fixtureExif{200, 100, "0/1"},
    &fixtureExif{200, 100, "-2/1"},
    &fixtureExif{200, 100, "2/1"},
  }

  errs := []error{
    nil,
    nil,
    nil,
  }

  Find(fixtureFileFinder{ files: files }, &fixtureDecoder{ exifs: exifs, errs: errs }, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}
