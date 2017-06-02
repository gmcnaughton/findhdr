package findhdr

/*
TODO
----
- what about configurable # of shots?
- allow any variation of bias values
- error handling, holy sweet jesus and mary - where do the errs go?
- should we really be treating bias value as a string? why didn't StringVal() work?
- support non-JPG filenames
- print out some stats afterwards
- configurable verbosity, please!
*/

import (
  "path/filepath"
  "os"
  "fmt"

  "github.com/rwcarlsen/goexif/exif"
)

type Hdr struct {
  a *Image
  b *Image
  c *Image
}

type Image struct {
  Path string
  Info os.FileInfo

  exif Exif
}

type Exif interface {
  PixelXDimension() (int, error)
  PixelYDimension() (int, error)
  ExposureBiasValue() (string, error)
}

/*
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

func (wrapper *exifWrapper) PixelXDimension() (val int, err error) {
  tag, err := wrapper.exif.Get(exif.PixelXDimension)
  if err != nil {
    return 0, err
  }

  val, err = tag.Int(0)
  if err != nil {
    return 0, err
  }

  return
}

func (wrapper *exifWrapper) PixelYDimension() (val int, err error) {
  tag, err := wrapper.exif.Get(exif.PixelYDimension)
  if err != nil {
    return 0, err
  }

  val, err = tag.Int(0)
  if err != nil {
    return 0, err
  }

  return
}

func (wrapper *exifWrapper) ExposureBiasValue() (val string, err error) {
  tag, err := wrapper.exif.Get(exif.ExposureBiasValue)
  if err != nil {
    return "", err
  }

  val = tag.String()
  return
}

func (hdr *Hdr) Add(x Exif, path string, info os.FileInfo) {
  img := &Image{ path, info, x }
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

func (hdr *Hdr) IsHdr() bool {
  if hdr.a == nil || hdr.b == nil || hdr.c == nil {
    // fmt.Println("Skipping: insufficient candidates")
    return false
  }

  ay, _ := hdr.a.exif.PixelYDimension()
  by, _ := hdr.b.exif.PixelYDimension()
  cy, _ := hdr.c.exif.PixelYDimension()

  ax, _ := hdr.a.exif.PixelXDimension()
  bx, _ := hdr.b.exif.PixelXDimension()
  cx, _ := hdr.c.exif.PixelXDimension()

  abias, _ := hdr.a.exif.ExposureBiasValue()
  bbias, _ := hdr.b.exif.ExposureBiasValue()
  cbias, _ := hdr.c.exif.ExposureBiasValue()

  if ax != bx || bx != cx {
    // fmt.Println("Skipping: x dimension mismatch", ax, bx, cx)
    return false
  }

  if ay != by || by != cy {
    // fmt.Println("Skipping: y dimension mismatch", ay, by, cy)
    return false
  }

  if abias != "\"0/1\"" || bbias != "\"-2/1\"" || cbias != "\"2/1\"" {
    // fmt.Println("Skipping: bias mismatch", abias, bbias, cbias)
    return false
  }

  return true
}

func (hdr *Hdr) Images() []*Image {
  return []*Image{
    hdr.a,
    hdr.b,
    hdr.c,
  }
}

type FileFinderFunc filepath.WalkFunc

type FileFinder interface {
  Find(fileFinderFn FileFinderFunc)
}

type FilePathWalker struct {
  Root string
}

func (f FilePathWalker) Find(fileFinderFn FileFinderFunc) {
  filepath.Walk(f.Root, func(path string, info os.FileInfo, err error) error {
    return fileFinderFn(path, info, err)
  })
}

type Decoder interface {
  Decode(path string) (exif Exif, err error)
}

type ExifDecoder struct {}

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

type HdrFunc func(hdr *Hdr)

func Find(finder FileFinder, decoder Decoder, hdrFn HdrFunc) {
  hdr := Hdr{}

  // See https://golang.org/pkg/path/filepath/#WalkFunc
  finder.Find(func(path string, info os.FileInfo, err error) error {
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
    if hdr.IsHdr() {
      if hdrFn != nil {
        hdrFn(&hdr)
      }
      hdr = Hdr{}
    }

    return nil // or SkipDir to skip processng this dir
  })
}
