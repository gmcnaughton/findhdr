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
	isHdr, err := hdr.IsHdr()
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if isHdr {
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

func (f testFileFinder) Find(fileFinderFn FileFinderFunc) error {
	if f.err != nil {
		return fileFinderFn("", nil, f.err)
	}

	for _, f := range f.files {
		if err := fileFinderFn(f.path, f.info, f.err); err != nil {
			return err
		}
	}
	return nil
}

func TestFindNonExistantDirectory(t *testing.T) {
	err := Find(testFileFinder{err: os.ErrNotExist}, NewDecoder(), func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
	if err != os.ErrNotExist {
		t.Error("Expected Find to return error")
	}
}

func TestFindEmptyDirectory(t *testing.T) {
	err := Find(testFileFinder{}, NewDecoder(), func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
	if err != nil {
		t.Error("Expected no error")
	}
}

func TestFindNonImageFiles(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.txt"},
		testFile{path: "foo2.txt"},
		testFile{path: "foo3.txt"},
	}

	err := Find(testFileFinder{files: files}, NewDecoder(), func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
	if err != nil {
		t.Error("Expected no error")
	}
}

func TestFindNonExistantImageFiles(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.JPG"},
		testFile{path: "foo2.JPG"},
		testFile{path: "foo3.JPG"},
	}

	err := Find(testFileFinder{files: files}, NewDecoder(), func(hdr *Hdr) {
		t.Error("Expected no HDRs to be found")
	})
	if err != nil {
		// TODO: should this be an error?
		t.Error("Expected no error")
	}
}

type testDecoder struct {
	metas []ImageMeta
	errs  []error
	calls int
}

// make sure it satisfies the interface
var _ Decoder = (*testDecoder)(nil)

func (decoder *testDecoder) Decode(path string) (meta ImageMeta, err error) {
	decoder.calls++
	return decoder.metas[decoder.calls-1], decoder.errs[decoder.calls-1]
}

type testImageMeta struct {
	xdim int
	ydim int
	bias string
}

// make sure it satisfies the interface
var _ ImageMeta = (*testImageMeta)(nil)

func (m *testImageMeta) PixelXDimension() (val int, err error) {
	return m.xdim, nil
}

func (m *testImageMeta) PixelYDimension() (val int, err error) {
	return m.ydim, nil
}

func (m *testImageMeta) ExposureBiasValue() (val string, err error) {
	return fmt.Sprintf("\"%s\"", m.bias), nil
}

func TestFindSuccess(t *testing.T) {
	files := []testFile{
		testFile{path: "foo1.JPG"},
		testFile{path: "foo2.JPG"},
		testFile{path: "foo3.JPG"},
	}

	metas := []ImageMeta{
		&testImageMeta{200, 100, "0/1"},
		&testImageMeta{200, 100, "-2/1"},
		&testImageMeta{200, 100, "2/1"},
	}

	errs := []error{
		nil,
		nil,
		nil,
	}

	called := 0
	err := Find(testFileFinder{files: files}, &testDecoder{metas: metas, errs: errs}, func(hdr *Hdr) {
		called++
	})
	if err != nil {
		t.Error("Expected no error")
	}
	if called != 1 {
		t.Errorf("Expected 1 HDR to be found but got %d", called)
	}
}
