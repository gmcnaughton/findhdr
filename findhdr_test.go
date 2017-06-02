package findhdr

import (
	"fmt"
	"os"
	"testing"
)

// Usage:
//     go test -v ./...

func TestHdrIsHdrWhenEmpty(t *testing.T) {
	hdr := Hdr{}
	if hdr.IsHdr() {
		t.Error("Expected empty Hdr not to be IsHdr()")
	}
}

type testFile struct {
	path string
	info os.FileInfo
	err  error
}

type testFileFinder struct {
	files []testFile
	err   error
}

// make sure it satisfies the interface
var _ FileFinder = (*testFileFinder)(nil)

func (f testFileFinder) Find(fileFinderFn FileFinderFunc) {
	if f.err != nil {
		fileFinderFn("", nil, f.err)
	} else {
		for _, f := range f.files {
			fileFinderFn(f.path, f.info, f.err)
		}
	}
}

func TestFindNonExistantDirectory(t *testing.T) {
	Find(testFileFinder{err: os.ErrNotExist}, &ExifDecoder{}, func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
}

func TestFindEmptyDirectory(t *testing.T) {
	Find(testFileFinder{}, &ExifDecoder{}, func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
}

func TestFindNonImageFiles(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.txt"},
		testFile{path: "foo2.txt"},
		testFile{path: "foo3.txt"},
	}

	Find(testFileFinder{files: files}, &ExifDecoder{}, func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
}

func TestFindNonExistantImageFiles(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.JPG"},
		testFile{path: "foo2.JPG"},
		testFile{path: "foo3.JPG"},
	}

	Find(testFileFinder{files: files}, &ExifDecoder{}, func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
}

type testDecoder struct {
	exifs []Exif
	errs  []error
	calls int
}

// make sure it satisfies the interface
var _ Decoder = (*testDecoder)(nil)

func (decoder *testDecoder) Decode(path string) (exif Exif, err error) {
	decoder.calls++
	return decoder.exifs[decoder.calls-1], decoder.errs[decoder.calls-1]
}

type testExif struct {
	xdim int
	ydim int
	bias string
}

// make sure it satisfies the interface
var _ Exif = (*testExif)(nil)

func (exif *testExif) PixelXDimension() (val int, err error) {
	return exif.xdim, nil
}

func (exif *testExif) PixelYDimension() (val int, err error) {
	return exif.ydim, nil
}

func (exif *testExif) ExposureBiasValue() (val string, err error) {
	return fmt.Sprintf("\"%s\"", exif.bias), nil
}

func TestFindSuccess(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.JPG"},
		testFile{path: "foo2.JPG"},
		testFile{path: "foo3.JPG"},
	}

	exifs := []Exif{
		&testExif{200, 100, "0/1"},
		&testExif{200, 100, "-2/1"},
		&testExif{200, 100, "2/1"},
	}

	errs := []error{
		nil,
		nil,
		nil,
	}

	called := 0
	Find(testFileFinder{files: files}, &testDecoder{exifs: exifs, errs: errs}, func(hdr *Hdr) {
		called += 1
	})
	if called != 1 {
		t.Errorf("Expected 1 HDR to be found but got %d", called)
	}
}
