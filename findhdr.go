package findhdr

/*
TODO
----
- should we really be treating bias value as a string? why didn't StringVal() work?
- print out some stats afterwards
- verbose mode!
*/

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gmcnaughton/go-experiments/circbuf"
	"github.com/xor-gate/goexif2/exif"
)

func init() {
	addMimeTypes()
}

// Hdr represents a set of candidate images which may -- or may not -- represent
// valid sources for a single merged hdr image. See IsHdr() for details on
// what constitues valid sources.
type Hdr struct {
	images *circbuf.Circbuf
	min    int
	max    int
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

// NewHdr returns a reference to an initialized Hdr candidate struct.
func NewHdr(min int, max int) *Hdr {
	return &Hdr{
		images: circbuf.New(max),
		min:    min,
		max:    max,
	}
}

// Add adds a candidate image to an Hdr.
func (hdr *Hdr) Add(meta ImageMeta, path string, info os.FileInfo) {
	img := &Image{Path: path, Info: info, Meta: meta}
	hdr.images.Add(img)
}

func (hdr *Hdr) String() string {
	names := []string{}
	err := hdr.images.Do(func(item interface{}) error {
		img, ok := item.(*Image)
		if !ok {
			return fmt.Errorf("unrecognized type (not Image): %T", item)
		}
		names = append(names, img.Info.Name())
		return nil
	})
	if err != nil {
		return fmt.Sprintf("[error: %v]", err)
	}
	return fmt.Sprintf("%v", names)
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
	return hdr.images.Len() >= hdr.min
}

func (hdr *Hdr) dimensionsMatch() (bool, error) {
	ydim, xdim := 0, 0
	match := true
	err := hdr.images.Do(func(item interface{}) error {
		img, ok := item.(*Image)
		if !ok {
			return fmt.Errorf("invalid type: %T", item)
		}

		y, err := img.Meta.PixelYDimension()
		if err != nil {
			match = false
			return err
		}
		x, err := img.Meta.PixelXDimension()
		if err != nil {
			match = false
			return err
		}
		if ydim == 0 {
			ydim = y
		} else if y != ydim {
			match = false
			return err
		}
		if xdim == 0 {
			xdim = x
		} else if x != xdim {
			match = false
			return err
		}
		return nil
	})
	return match, err
}

func (hdr *Hdr) biasValuesUnique() (bool, error) {
	biases := map[string]bool{}
	unique := true
	err := hdr.images.Do(func(item interface{}) error {
		img, ok := item.(*Image)
		if !ok {
			return fmt.Errorf("invalid type: %T", item)
		}

		bias, err := img.Meta.ExposureBiasValue()
		if err != nil {
			unique = false
			return err
		}

		if _, ok := biases[bias]; ok {
			unique = false
			return err
		}
		biases[bias] = true
		return nil
	})
	return unique, err
}

// Images returns the candidate images in this Hdr.
func (hdr *Hdr) Images() []*Image {
	images := make([]*Image, 0, hdr.images.Len())
	err := hdr.images.Do(func(item interface{}) error {
		img, ok := item.(*Image)
		if !ok {
			return fmt.Errorf("invalid type: %T", item)
		}

		images = append(images, img)
		return nil
	})
	if err != nil {
		panic("runtime error: findhdr.Hdr.Images: error while iterating")
	}
	return images
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
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			panic("runtime error: findhdr.exifDecoder.Decode: error while decoding")
		}
	}()

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
func Find(finder FileFinder, decoder Decoder, min int, max int, hdrFn HdrFunc) error {
	hdr := NewHdr(min, max)

	// See https://golang.org/pkg/path/filepath/#WalkFunc
	return finder.Find(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info != nil && info.IsDir() {
			return nil
		}

		if isimage := isImageFile(path); !isimage {
			return nil
		}

		x, err := decoder.Decode(path)
		if err != nil {
			// TODO: in verbose mode, warn about files we tried to parse but couldn't?
			// fmt.Println(err)
			return nil
		}

		hdr.Add(x, path, info)

		isHdr, err := hdr.IsHdr()
		if err != nil {
			return err
		}

		if isHdr {
			if hdrFn != nil {
				hdrFn(hdr)
			}
			hdr = NewHdr(hdr.min, hdr.max)
		}

		return nil // or SkipDir to skip processng this dir
	})
}

// isImageFile returns true if the image at the given path has a mime-type of
// "image/*". Mime types are calculated using mime.TypeByExtension, which
// understands mime types built in to the system, as well as custom mime types
// mappings which we have added for all known RAW image extensions.
func isImageFile(path string) bool {
	ext := filepath.Ext(path)
	mimetype := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimetype, "image/")
}

func addMimeTypes() {
	// See https://bugs.freedesktop.org/show_bug.cgi?id=8170
	var mimeTypes = []struct {
		ext      string
		mimetype string
	}{
		{".crw", "image/x-canon-crw"},
	}

	for _, mimetype := range mimeTypes {
		err := mime.AddExtensionType(mimetype.ext, mimetype.mimetype)
		if err != nil {
			fmt.Println(err)
		}
	}
}
