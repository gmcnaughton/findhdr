package findhdr

/*
- what about configurable # of shots?
- allow any variation of bias values
- ignore non-image file extensions
- error handling, holy sweet jesus and mary - where do the errs go?
- should we really be treating bias value as a string? why didn't StringVal() work?
- support non-JPG filenames
- print out some stats afterwards
- configurable verbosity, please!
- configuration that just prints out matching file names to stdout instead of hardlinking
- use the findhdr.Image abstraction internally

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

import (
  "path/filepath"
  "os"
  "fmt"

  "github.com/rwcarlsen/goexif/exif"
)

type Hdr struct {
  a *exif.Exif
  b *exif.Exif
  c *exif.Exif

  ai os.FileInfo
  bi os.FileInfo
  ci os.FileInfo

  ap string
  bp string
  cp string
}

type Image struct {
  Info os.FileInfo
  Path string
}

func (hdr *Hdr) Add(x *exif.Exif, path string, info os.FileInfo) {
  if hdr.a == nil {
    hdr.a = x
    hdr.ap = path
    hdr.ai = info
  } else if hdr.b == nil {
    hdr.b = x
    hdr.bp = path
    hdr.bi = info
  } else if hdr.c == nil {
    hdr.c = x
    hdr.cp = path
    hdr.ci = info
  } else {
    hdr.a = hdr.b
    hdr.ap = hdr.bp
    hdr.ai = hdr.bi
    hdr.b = hdr.c
    hdr.bp = hdr.cp
    hdr.bi = hdr.ci
    hdr.c = x
    hdr.cp = path
    hdr.ci = info
  }
}

func (hdr Hdr) String() string {
  return fmt.Sprintf("HDR [%s, %s, %s]", hdr.ai.Name(), hdr.bi.Name(), hdr.ci.Name())
}

func (hdr *Hdr) IsHdr() bool {
  if hdr.a == nil || hdr.b == nil || hdr.c == nil {
    // fmt.Println("Skipping: insufficient candidates")
    return false
  }

  aytag, _ := hdr.a.Get(exif.PixelYDimension)
  bytag, _ := hdr.b.Get(exif.PixelYDimension)
  cytag, _ := hdr.c.Get(exif.PixelYDimension)

  axtag, _ := hdr.a.Get(exif.PixelXDimension)
  bxtag, _ := hdr.b.Get(exif.PixelXDimension)
  cxtag, _ := hdr.c.Get(exif.PixelXDimension)

  abiastag, _ := hdr.a.Get(exif.ExposureBiasValue)
  bbiastag, _ := hdr.b.Get(exif.ExposureBiasValue)
  cbiastag, _ := hdr.c.Get(exif.ExposureBiasValue)

  ax, _ := axtag.Int(0)
  bx, _ := bxtag.Int(0)
  cx, _ := cxtag.Int(0)

  ay, _ := aytag.Int(0)
  by, _ := bytag.Int(0)
  cy, _ := cytag.Int(0)

  abias := abiastag.String()
  bbias := bbiastag.String()
  cbias := cbiastag.String()

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

func (hdr *Hdr) Images() []Image {
  return []Image{
    Image{ Path: hdr.ap, Info: hdr.ai },
    Image{ Path: hdr.bp, Info: hdr.bi },
    Image{ Path: hdr.cp, Info: hdr.ci },
  }
}

func Find(root string, findfn func(hdr *Hdr)) {
  hdr := Hdr{}

  // See https://golang.org/pkg/path/filepath/#WalkFunc
  filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if info.IsDir() {
      return nil
    }

    ext := filepath.Ext(path)
    if ext != ".JPG" {
      return nil
    }

    f, err := os.Open(path)
    if err != nil {
        fmt.Println(err)
        return nil
    }
    defer f.Close()

    x, err := exif.Decode(f)
    if err != nil {
        fmt.Println(err)
        return nil
    }

    hdr.Add(x, path, info)
    if hdr.IsHdr() {
      if findfn != nil {
        findfn(&hdr)
      }
      hdr = Hdr{}
    }

    return nil // or SkipDir to skip processng this dir
  })
}
