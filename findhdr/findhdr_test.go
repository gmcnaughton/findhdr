package findhdr

import (
  "testing"
  "os"
)

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
  Find(fixtureFileFinder{ err: os.ErrNotExist }, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindEmptyDirectory(t *testing.T) {
  Find(fixtureFileFinder{ }, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindNonImageFiles(t *testing.T) {
  files := []fixtureFile{
    fixtureFile{ path: "foo1.txt" },
    fixtureFile{ path: "foo2.txt" },
    fixtureFile{ path: "foo3.txt" },
  }

  Find(fixtureFileFinder{ files: files }, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}

func TestFindNonExistantImageFiles(t *testing.T) {
  files := []fixtureFile{
    fixtureFile{ path: "foo1.JPG" },
    fixtureFile{ path: "foo2.JPG" },
    fixtureFile{ path: "foo3.JPG" },
  }

  Find(fixtureFileFinder{ files: files }, func(hdr *Hdr) {
    t.Error("Expected no HDRs to get reported")
  })
}
