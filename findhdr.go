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

	exif Exif
}

// Exif is an interface providing access to image metadata.
type Exif interface {
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
type exifWrapper struct {
	exif *exif.Exif
}

// PixelXDimension returns the width of the image.
func (wrapper *exifWrapper) PixelXDimension() (val int, err error) {
	tag, err := wrapper.exif.Get(exif.PixelXDimension)
	if err != nil {
		return
	}

	val, err = tag.Int(0)
	return
}

// PixelYDimension returns the height of the image.
func (wrapper *exifWrapper) PixelYDimension() (val int, err error) {
	tag, err := wrapper.exif.Get(exif.PixelYDimension)
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
func (wrapper *exifWrapper) ExposureBiasValue() (val string, err error) {
	tag, err := wrapper.exif.Get(exif.ExposureBiasValue)
	if err != nil {
		return
	}

	val = tag.String()
	return
}

// Add adds a candidate image to an Hdr.
func (hdr *Hdr) Add(x Exif, path string, info os.FileInfo) {
	img := &Image{path, info, x}
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
	if hdr.a == nil || hdr.b == nil || hdr.c == nil {
		// fmt.Println("Skipping: insufficient candidates")
		return false, nil
	}

	ay, err := hdr.a.exif.PixelYDimension()
	if err != nil {
		return false, err
	}
	by, err := hdr.b.exif.PixelYDimension()
	if err != nil {
		return false, err
	}
	cy, err := hdr.c.exif.PixelYDimension()
	if err != nil {
		return false, err
	}

	ax, err := hdr.a.exif.PixelXDimension()
	if err != nil {
		return false, err
	}
	bx, err := hdr.b.exif.PixelXDimension()
	if err != nil {
		return false, err
	}
	cx, err := hdr.c.exif.PixelXDimension()
	if err != nil {
		return false, err
	}

	abias, err := hdr.a.exif.ExposureBiasValue()
	if err != nil {
		return false, err
	}
	bbias, err := hdr.b.exif.ExposureBiasValue()
	if err != nil {
		return false, err
	}
	cbias, err := hdr.c.exif.ExposureBiasValue()
	if err != nil {
		return false, err
	}

	if ax != bx || bx != cx {
		// fmt.Println("Skipping: x dimension mismatch", ax, bx, cx)
		return false, nil
	}

	if ay != by || by != cy {
		// fmt.Println("Skipping: y dimension mismatch", ay, by, cy)
		return false, nil
	}

	if abias != "\"0/1\"" || bbias != "\"-2/1\"" || cbias != "\"2/1\"" {
		// fmt.Println("Skipping: bias mismatch", abias, bbias, cbias)
		return false, nil
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

// FilePathWalker is a FileFinder using filepath.Walk to visit all files
// and directories in the given root path.
type FilePathWalker struct {
	Root string
}

// Find walks the file tree rooted at root, calling fileFinderFn for each file or
// directory in the tree, including root.
func (f FilePathWalker) Find(fileFinderFn FileFinderFunc) error {
	return filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
		return fileFinderFn(path, info, err)
	})
}

// Decoder is the interface providing exif metadata for an image file via the
// Decode method.
type Decoder interface {
	Decode(path string) (exif Exif, err error)
}

// ExifDecoder is a Decoder using exif.Exif to extract exif tags from an image file.
type ExifDecoder struct{}

// Decode returns Exif metadata for an image file. If the file does not exist,
// we do not have permission to read it, or if any other error occurs while
// extracting metadata from the image, it will be returned along with nil metadata.
func (d *ExifDecoder) Decode(path string) (Exif, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w := new(exifWrapper)
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
