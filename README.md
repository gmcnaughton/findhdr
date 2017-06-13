# Find Hdr

`findhdr` is a command-line tool for finding HDR source images in directories with a mix of HDR and non-HDR photos (such as your digital camera import after a vacation).

Images are considered HDR sources if they have contiguous file names (*IMG_0001.JPG, IMG_0002.JPG, IMG_003.JPG*), identical dimensions (*5184 x 3456*), and unique exposure bias values (*2/1, -2/1, 0/1*). This indicates a burst of images taken using auto exposure bracketing (AEB).

Images found by `findhdr` can be linked into a target folder, in preparation for batch processing by [Photomatix](https://www.hdrsoft.com/).

### Installation

    go get github.com/gmcnaughton/findhdr

### Usage

Print a list of HDR sources in a directory:

    findhdr /path/to/photos

Link HDR sources found into the `-link` directory, creating it if necessary:

    findhdr -link ./out /path/to/photos

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/gmcnaughton/findhdr.

## License

The tool is available as open source under the terms of the [MIT License](http://opensource.org/licenses/MIT).
