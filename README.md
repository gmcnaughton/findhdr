# Find Hdr

This repository contains a tool for finding HDR images in directories containing a mix of HDR and non-HDR images (such as imported photos from a digital camerap).

Images are considered part of an HDR if a series of images exist with contiguous file names (*IMG_0001.jpg, IMG_0002.jpg, IMG_003.jpg*) and identical dimensions (*1024x768*), but different exposure bias values (*+2/-2/0*).

Images found by `findhdr` can be linked into a target folder, to prepare for batch processing by [Photomatix](https://www.hdrsoft.com/).

### Installation

    go get github.com/gmcnaughton/findhdr

### Usage

Print a list of HDRs found in a directory:

    findhdr /path/to/photos

Link HDRs found into the `-link` directory:

    findhdr -link ./out /path/to/photos

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/gmcnaughton/findhdr.

## License

The gem is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).
