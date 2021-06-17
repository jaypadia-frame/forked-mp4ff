![Logo](images/logo.png)

![Test](https://github.com/jaypadia-frame/mp4ff/workflows/Go/badge.svg)
![golangci-lint](https://github.com/jaypadia-frame/mp4ff/workflows/golangci-lint/badge.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/jaypadia-frame/mp4ff?status.svg)](http://godoc.org/github.com/jaypadia-frame/mp4ff)
[![Go Report Card](https://goreportcard.com/badge/github.com/jaypadia-frame/mp4ff)](https://goreportcard.com/report/github.com/jaypadia-frame/mp4ff)
[![license](https://img.shields.io/github/license/jaypadia-frame/mp4ff.svg)](https://github.com/jaypadia-frame/mp4ff/blob/master/LICENSE.md)

**NOTE: this is a fork from edgeware/mp4ff - used only for experimental purpose right now**

Package mp4ff implements MP4 media file parser and writer for AVC and HEVC video, AAC audio and stpp/wvtt subtitles.
Focused on fragmented files as used for streaming in DASH, MSS and HLS fMP4.

## Library

The library has functions for parsing (called Decode) and writing (Encode) in the package `mp4ff/mp4`.
It also contains codec specific parsing of of AVC/H.264 including complete parsing of
SPS and PPS in the package `mp4ff.avc`. HEVC/H.265 parsing is less complete, and available as `mp4ff.hevc`.

Traditional multiplexed non-fragmented mp4 files can also be parsed and decoded, see `examples/segment`.

The focus is, however, on non-multiplexed single-track fragmented mp4 files as used in DASH, HLS, and CMAF.

The top level structure for both non-fragmented and fragmented mp4 files is `mp4.File`.

In the non-fragmented files, the members Ftyp, Moov, and Mdat are used.
A fragmented `mp4.File` file can be a single init segment, one or more media segments, or a a
combination of both like a CMAF track which renders into a playable one-track asset.

For fragmented files, following high-level structures are used:

* `InitSegment` contains an `ftyp` and `moov` box and provides the metadata for a fragmented files.
   It corresponds to a CMAF header
* `MediaSegment` starts with an optional `styp` box and contains on or more `Fragment`s
* `Fragment` is an mp4 fragment with exactly one `moof` box followed by a `mdat` box where the latter
   contains the media data. It is limited to have exactly one `trun` box.

The typical child boxes are exported so that one can write paths such as

    fragment.Moof.Traf.Trun

to access the (only) trun box in a fragment.

The codecs currently supported are AVC/H.264, HEVC/H.265 and AAC.

## Usage for creating new fragmented files

A typical use case is to produce an init segment and followed by a series of media segments.

The first step is to create an init segment. This is done in three steps and can be seen in
`examples/initcreator:

1. `init := mp4.CreateEmptyInit()`
2. `init.AddEmptyTrack(timescale, mediatype, language)`
3. `init.Moov.Trak.SetHEVCDescriptor("hvc1", vpsNALUs, spsNALUs, ppsNALUs)`

where the 3'rd step fills in codec-specific parameters into the sample descriptor of the first track.
Multiple tracks are also available as via the slice variable `Traks` instead of `Trak`.

The second step is to start producing media segments. They should use the timescale that
was set when creating the init segments. The timescales should be chosen so that the
sample durations have exact values.

A media segment contains one or more fragments, where each fragment has a moof and an mdat box.
If all samples are available before the segment is to be used, one can use use a single
fragment in each segment. Example code for this can be found in `examples/segmenter`.

The high-level code to do that is to first create a slice of `FullSample` with the data needed.
All times are in the track timescale set when creating the init segment and coded in the `mdhd` box.

	mp4.FullSample{
		Sample: mp4.Sample{
	        Flags uint32 // Flag sync sample etc
	        Dur   uint32 // Sample duration in mdhd timescale
	        Size  uint32 // Size of sample data
	        Cto   int32  // Signed composition time offset
		},
	    DecodeTime uint64 // Absolute decode time (offset + accumulated sample Dur)
	    Data       []byte // Sample data
	}

The `mp4.Sample` part is what will be written into the `trun` box.
`DecodeTime` is the media timeline accumulated time. The value of the first samples of a fragment, will
be set as the `BaseMediaDecodeTime` in the `tfdt` box when calling Encode on the fragment.
Once a number of such samples are available, then can be added to a media segment

	seg := mp4.NewMediaSegment()
	frag := mp4.CreateFragment(uint32(segNr), mp4.DefaultTrakID)
	seg.AddFragment(frag)
	for _, sample := range samples {
		frag.AddSample(sample)
	}

This segment can finally be output to a `w io.Writer` as

    err := seg.Encode(w)

## Direct changes of attributes

Many attributes are public and can therefore be changed in an uncontrolled way.
The advantage of this is that it is possible to write code that can manipulate boxes
in many different ways. However, some attributes and linkd can become broken.

As an example, container boxes such as `TrafBox` have a method `AddChild` which
adds a box to its slice of children boxes `Children`, but also sets a specific
member reference such as `Tfdt`  to point to that box. If `Children` is manipulated
directly, that link will not be valid.

## Encoding modes and optimizations
For fragmented files, one can choose to either encode all boxes in a file, or only code
the ones which are inclueded in the init and media segments. The attribute that controls that
is called `FragEncMode`.
Another attribute `EncOptimize` controls possible optimizations of the file encoding process.
Currently there is only one possible optimization called `OptimizeTrun`.
It can reduce the size of the `TrunBox` by finding and writing default
values in the `TfhdBox` and omitting the corresponding values from the `TrunBox`.
Note that this may change the size of all ancestor boxes of `trun`.

## Sample Number Offset
Following the ISOBMFF standard, sample numbers and other numbers start at 1 (one-based).
This applies to arguments of functions. The actual storage in slices are zero-based, so
sample nr 1 has index 0 in the corresponding slice.

## Command Line Tools

Some useful command line tools are available in `cmd`.

1. `mp4ff-info` prints a tree of the box hierarchy of an mp4 file with information
    of the boxes. The level of detail can be increased with the option `-l`, like `-l all:1` for all boxes or `-l trun:1,stss:1` for specific boxes.
2. `mp4ff-pslister` extracts and displays pps and sps for AVC in an mp4 file.
3. `mp4ff-nallister` lists NALus and picture types for video in progressive or fragmented file
4. `mp4ff-wvttlister` lists details of wvtt (WebVTT in ISOBMFF) samples

You can install them by going to their respective directory and run `go install .`.

## Example code

Example code is available in the `examples` directory.
These are

1. `initcreator` which creates typical init segments (ftyp + moov) for video and audio
2. `resegmenter` which reads a segmented file (CMAF track) and resegments it with other
    segment durations.
3. `segmenter` which takes a progressive mp4 file and creates init and media segments from it.
4. `multitrack` parses a fragmented file with multiple tracks

## Stability
The APIs should be fairly stable, but minor non-backwards-compatible tweaks may happen until version 1.

## Specifications
The main specification for the MP4 file format is the ISO Base Media File Format (ISOBMFF) standard
ISO/IEC 14496-12 6'th ed 2020. Some boxes are specified in other standards, as should be commented
in the code.

## LICENSE

MIT, see [LICENSE.md](LICENSE.md).

Some code in pkg/mp4, comes from or is based on https://github.com/jfbus/mp4 which has
`Copyright (c) 2015 Jean-François Bustarret`.

Some code in pkg/bits comes from or is based on https://github.com/tcnksm/go-casper/tree/master/internal/bits
`Copyright (c) 2017 Taichi Nakashima`.

## Versions

See [Versions.md](Versions.md).
