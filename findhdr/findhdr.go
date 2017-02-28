package findhdr

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

import (
  "path/filepath"
  "os"
  "fmt"

  "github.com/rwcarlsen/goexif/exif"
)

func Find(root string) {
  // See https://golang.org/pkg/path/filepath/#WalkFunc
  filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if !info.IsDir() {
      fmt.Println(info.Name())

      f, err := os.Open(path)
      if err != nil {
          fmt.Println(err)
          return nil
      }
      defer f.Close()

      x, err := exif.Decode(f) // exif.Exif
      if err != nil {
          fmt.Println(err)
          return nil
      }

      // PixelYDimension, PixelXDimension, ExposureBiasValue
      bias, _ := x.Get(exif.ExposureBiasValue)
      fmt.Printf("%s = %s\n", path, bias) // "0/1", "-2/1", "2/1"
    }

    return nil // or SkipDir to skip processng this dir
  })
}
