package mp4

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestMediaSegmentFragmentation(t *testing.T) {

	trex := &TrexBox{
		TrackID: 2,
	}

	inFile := "testdata/1.m4s"
	inFileGoldenDumpPath := "testdata/golden_1_m4s_dump.txt"
	goldenFragPath := "testdata/golden_1_frag.m4s"
	goldenFragDumpPath := "testdata/golden_1_frag_m4s_dump.txt"
	fd, err := os.Open(inFile)
	if err != nil {
		t.Error(err)
	}
	defer fd.Close()

	f, err := DecodeFile(fd)
	if err != io.EOF && err != nil {
		t.Error(err)
	}
	if len(f.Segments) != 1 {
		t.Errorf("Not exactly one mediasegment")
	}

	var bufInSeg bytes.Buffer
	f.EncOptimize = OptimizeNone // Avoid trun optimization
	f.FragEncMode = EncModeBoxTree
	err = f.Encode(&bufInSeg)
	if err != nil {
		t.Error(err)
	}

	inSeg, err := ioutil.ReadFile(inFile)
	if err != nil {
		t.Error(err)
	}

	diff := deep.Equal(inSeg, bufInSeg.Bytes())
	if diff != nil {
		t.Errorf("Written segment differs from %s", inFile)
	}

	err = compareOrUpdateInfo(t, f, inFileGoldenDumpPath)

	if err != nil {
		t.Error(err)
	}

	mediaSegment := f.Segments[0]
	var timeScale uint64 = 90000
	var duration uint32 = 45000

	fragments, err := mediaSegment.Fragmentify(timeScale, trex, duration)
	if err != nil {
		t.Errorf("Fragmentation went wrong")
	}
	if len(fragments) != 4 {
		t.Errorf("%d fragments instead of 4", len(fragments))
	}

	var bufFrag bytes.Buffer
	fragmentedSegment := NewMediaSegment()
	fragmentedSegment.EncOptimize = OptimizeTrun
	fragmentedSegment.Styp = f.Segments[0].Styp
	fragmentedSegment.Fragments = fragments

	err = fragmentedSegment.Encode(&bufFrag)
	if err != nil {
		t.Error(err)
	}

	err = compareOrUpdateInfo(t, fragmentedSegment, goldenFragDumpPath)
	if err != nil {
		t.Error(err)
	}

	if *update {
		err = writeGolden(t, goldenFragPath, bufFrag.Bytes())
		if err != nil {
			t.Error(err)
		}
	} else {
		goldenFrag, err := ioutil.ReadFile(goldenFragPath)
		if err != nil {
			t.Error(err)
		}
		diff := deep.Equal(goldenFrag, bufFrag.Bytes())
		if diff != nil {
			t.Errorf("Generated dump different from %s", goldenFragPath)
		}
	}
}

func TestDoubleDecodeEncodeOptimize(t *testing.T) {
	inFile := "testdata/1.m4s"

	fd, err := os.Open(inFile)
	if err != nil {
		t.Error(err)
	}
	defer fd.Close()

	enc1 := decodeEncode(t, fd, OptimizeTrun)
	buf1 := bytes.NewBuffer(enc1)
	enc2 := decodeEncode(t, buf1, OptimizeTrun)
	diff := deep.Equal(enc2, enc1)
	if diff != nil {
		t.Errorf("Second write gives diff %s", diff)
	}
}

func TestDecodeEncodeNoOptimize(t *testing.T) {

	inFile := "testdata/1.m4s"

	data, err := ioutil.ReadFile(inFile)
	if err != nil {
		t.Error(err)
	}
	buf0 := bytes.NewBuffer(data)
	enc := decodeEncode(t, buf0, OptimizeNone)
	diff := deep.Equal(enc, data)
	if diff != nil {
		t.Errorf("First encode gives diff %s", diff)
	}
}

func decodeEncode(t *testing.T, r io.Reader, optimize EncOptimize) []byte {
	f, err := DecodeFile(r)
	if err != nil {
		t.Error(err)
	}

	buf := bytes.Buffer{}
	f.EncOptimize = optimize
	err = f.Encode(&buf)
	if err != nil {
		t.Error(err)
	}
	return buf.Bytes()
}

func TestMoofEncrypted(t *testing.T) {

	inFile := "testdata/moof_enc.m4s"
	inFileGoldenDumpPath := "testdata/golden_moof_enc_m4s_dump.txt"
	fd, err := os.Open(inFile)
	if err != nil {
		t.Error(err)
	}
	defer fd.Close()

	f, err := DecodeFile(fd)
	if err != io.EOF && err != nil {
		t.Error(err)
	}

	var bufOut bytes.Buffer
	f.FragEncMode = EncModeBoxTree
	err = f.Encode(&bufOut)
	if err != nil {
		t.Error(err)
	}

	inSeg, err := ioutil.ReadFile(inFile)
	if err != nil {
		t.Error(err)
	}

	diff := deep.Equal(inSeg, bufOut.Bytes())
	if diff != nil {
		tmpOutput := "testdata/moof_enc_tmp.mp4"
		err := writeGolden(t, tmpOutput, bufOut.Bytes())
		if err == nil {
			t.Errorf("Encoded output not same as input for %s. Wrote %s", inFile, tmpOutput)
		} else {
			t.Errorf("Encoded output not same as input for %s, but error %s when writing  %s", inFile, err, tmpOutput)
		}
	}

	err = compareOrUpdateInfo(t, f, inFileGoldenDumpPath)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkDecodeEncode(b *testing.B) {
	inFile := "testdata/1.m4s"
	raw, _ := ioutil.ReadFile(inFile)

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(raw)
		f, _ := DecodeFile(buf)
		var bufInSeg bytes.Buffer
		f.FragEncMode = EncModeBoxTree
		_ = f.Encode(&bufInSeg)
	}
}
