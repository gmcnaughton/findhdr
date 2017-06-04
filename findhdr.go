package findhdr

/*
TODO
----
- what about configurable # of shots?
- allow any variation of bias values
- should we really be treating bias value as a string? why didn't StringVal() work?
- support non-JPG filenames
- print out some stats afterwards
- configurable verbosity!
*/

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
)

// Hdr represents a set of candidate images which may -- or may not -- represent
// valid sources for a single merged hdr image. See IsHdr() for details on
// what constitues valid sources.
type Hdr struct {
	a *Image
	b *Image
	c *Image
}

// Image represents a single image file and its metadata.
type Image struct {
	Path string
	Info os.FileInfo
	Meta ImageMeta
}

// ImageMeta is an interface providing access to image metadata.
type ImageMeta interface {
	PixelXDimension() (int, error)
	PixelYDimension() (int, error)
	ExposureBiasValue() (string, error)
}

/*
Metadata available from exif.Exif:

*exif.Exif, DateTime: "2017:02:26 13:04:32"
ExifVersion: "0221"
DateTimeOriginal: "2017:02:26 13:04:32"
ComponentsConfiguration: ""
Flash: 16
Make: "Canon"
ThumbJPEGInterchangeFormat: 10988
ExposureTime: "3/10"
ExposureProgram: 3
ExposureBiasValue: "0/1"
FocalLength: "18/1"
MakerNote: ""
SubSecTimeDigitized: "00"
ExifIFDPointer: 360
FocalPlaneXResolution: "5184000/907"
InteroperabilityIFDPointer: 9052
YResolution: "72/1"
UserComment: ""
PixelXDimension: 5184
FocalPlaneResolutionUnit: 2
ExposureMode: 0
GPSVersionID: [2,2,0,0]
Orientation: 8
WhiteBalance: 0
InteroperabilityIndex: "R98"
Artist: ""
FNumber: "35/10"
MeteringMode: 2
ColorSpace: 1
CustomRendered: 0
ThumbJPEGInterchangeFormatLength: 20838
YCbCrPositioning: 2
ApertureValue: "237568/65536"
SceneCaptureType: 0
Model: "Canon EOS 7D"
DateTimeDigitized: "2017:02:26 13:04:32"
SubSecTimeOriginal: "00"
PixelYDimension: 3456
GPSInfoIFDPointer: 9098
ResolutionUnit: 2
Copyright: ""
ISOSpeedRatings: 400
ShutterSpeedValue: "106496/65536"
SubSecTime: "00"
FlashpixVersion: "0100"
FocalPlaneYResolution: "3456000/595"
XResolution: "72/1"
*/
type exifMeta struct {
	exif *exif.Exif
}

// PixelXDimension returns the width of the image.
func (e *exifMeta) PixelXDimension() (val int, err error) {
	tag, err := e.exif.Get(exif.PixelXDimension)
	if err != nil {
		return
	}

	val, err = tag.Int(0)
	return
}

// PixelYDimension returns the height of the image.
func (e *exifMeta) PixelYDimension() (val int, err error) {
	tag, err := e.exif.Get(exif.PixelYDimension)
	if err != nil {
		return
	}

	val, err = tag.Int(0)
	return
}

// ExposureBiasValue returns the exposure bias value (exposure compensation)
// for the image. This is a string in the form "X/1" (NOTE: including the double-quotes):
// - "0/1" for normal
// - "-2/1" for darker
// - "2/1" for lighter)
func (e *exifMeta) ExposureBiasValue() (val string, err error) {
	tag, err := e.exif.Get(exif.ExposureBiasValue)
	if err != nil {
		return
	}

	val = tag.String()
	return
}

// Add adds a candidate image to an Hdr.
func (hdr *Hdr) Add(meta ImageMeta, path string, info os.FileInfo) {
	img := &Image{Path: path, Info: info, Meta: meta}
	if hdr.a == nil {
		hdr.a = img
	} else if hdr.b == nil {
		hdr.b = img
	} else if hdr.c == nil {
		hdr.c = img
	} else {
		hdr.a = hdr.b
		hdr.b = hdr.c
		hdr.c = img
	}
}

func (hdr *Hdr) String() string {
	return fmt.Sprintf("[%s, %s, %s]", hdr.a.Info.Name(), hdr.b.Info.Name(), hdr.c.Info.Name())
}

// IsHdr returns true if the Hdr contains a valid set of source images which can be
// merged into a single hdr image. Source images must have identical dimensions
// but unique exposure bias values (e.g., "0/1", "-2/1", and "2/1").
func (hdr *Hdr) IsHdr() (bool, error) {
	sufficient := hdr.sufficientImages()
	if !sufficient {
		// fmt.Println("Skipping: insufficient candidates")
		return false, nil
	}

	if match, err := hdr.dimensionsMatch(); err != nil || !match {
		// fmt.Println("Skipping: x dimension mismatch", ax, bx, cx)
		return match, err
	}

	if unique, err := hdr.biasValuesUnique(); err != nil || !unique {
		// fmt.Println("Skipping: bias mismatch", abias, bbias, cbias)
		return unique, err
	}

	return true, nil
}

func (hdr *Hdr) sufficientImages() bool {
	return !(hdr.a == nil || hdr.b == nil || hdr.c == nil)
}

func (hdr *Hdr) dimensionsMatch() (bool, error) {
	ydim, xdim := 0, 0
	for _, img := range hdr.Images() {
		y, err := img.Meta.PixelYDimension()
		if err != nil {
			return false, err
		}
		x, err := img.Meta.PixelXDimension()
		if err != nil {
			return false, err
		}
		if ydim == 0 {
			ydim = y
		} else if y != ydim {
			return false, nil
		}
		if xdim == 0 {
			xdim = x
		} else if x != xdim {
			return false, nil
		}
	}

	return true, nil
}

func (hdr *Hdr) biasValuesUnique() (bool, error) {
	biases := map[string]bool{}
	for _, img := range hdr.Images() {
		bias, err := img.Meta.ExposureBiasValue()
		if err != nil {
			return false, err
		}

		if _, ok := biases[bias]; ok {
			return false, nil
		}
		biases[bias] = true
	}
	return true, nil
}

// Images returns the candidate images in this Hdr.
func (hdr *Hdr) Images() []*Image {
	return []*Image{
		hdr.a,
		hdr.b,
		hdr.c,
	}
}

// FileFinderFunc is the type of the function called for each file or directory
// visited by Find.
type FileFinderFunc filepath.WalkFunc

// FileFinder is the interface providing a list of files or directories to
// process via the Find method.
type FileFinder interface {
	Find(fileFinderFn FileFinderFunc) error
}

// filePathWalker is a FileFinder using filepath.Walk to visit all files
// and directories in the given root path.
type filePathWalker struct {
	Root string
}

// NewFileFinder returns a FileFinder which walks all files in the given directory.
func NewFileFinder(path string) FileFinder {
	return filePathWalker{Root: path}
}

// Find walks the file tree rooted at root, calling fileFinderFn for each file or
// directory in the tree, including root.
func (f filePathWalker) Find(fileFinderFn FileFinderFunc) error {
	return filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
		return fileFinderFn(path, info, err)
	})
}

// Decoder is the interface providing exif metadata for an image file via the
// Decode method.
type Decoder interface {
	Decode(path string) (meta ImageMeta, err error)
}

// exifDecoder is a Decoder using exif.Exif to extract exif tags from an image file.
type exifDecoder struct{}

// NewDecoder returns a Decoder using exif.Exif to extract exif tags from an image file.
func NewDecoder() Decoder {
	return &exifDecoder{}
}

// Decode returns Exif metadata for an image file. If the file does not exist,
// we do not have permission to read it, or if any other error occurs while
// extracting metadata from the image, it will be returned along with nil metadata.
func (d *exifDecoder) Decode(path string) (ImageMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w := new(exifMeta)
	w.exif, err = exif.Decode(f)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// HdrFunc is the type of the function called for each valid Hdr found by Find.
type HdrFunc func(hdr *Hdr)

// Find visits all files found by the file finder, tests if they are images,
// uses the decoder to extract their metadata, and adds them to a candidate Hdr.
// Every valid Hdr found this way is delivered to HdrFunc (see IsHdr() for
// details on what constitutes a valid hdr).
func Find(finder FileFinder, decoder Decoder, hdrFn HdrFunc) error {
	hdr := Hdr{}

	// See https://golang.org/pkg/path/filepath/#WalkFunc
	return finder.Find(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info != nil && info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".JPG" {
			return nil
		}

		x, err := decoder.Decode(path)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		hdr.Add(x, path, info)

		isHdr, err := hdr.IsHdr()
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if isHdr {
			if hdrFn != nil {
				hdrFn(&hdr)
			}
			hdr = Hdr{}
		}

		return nil // or SkipDir to skip processng this dir
	})
}
