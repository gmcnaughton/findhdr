# Find Hdr

This repository contains a tool for finding HDR images in directories which contain a mix of HDR and non-HDR images.

Images are considered part of an HDR if there is a series of images with contiguous file names (*IMG_0001.jpg, IMG_0002.jpg, IMG_003.jpg*) and identical dimensions (*1024x768*) but different exposure bias values (*+2/-2/0*).

HDR images can be linked into a target folder in preparation for batch processing by [Photomatix](https://www.hdrsoft.com/).

### Getting Started

    go get github.com/gmcnaughton/findhdr

Print a list of HDRs found in a directory:

    findhdr -in ~/path/to/photos

Hard-link HDRs found in the `-in` directory to the `-out` directory:

    findhdr -in ~/path/to/photos -out ./out -link
