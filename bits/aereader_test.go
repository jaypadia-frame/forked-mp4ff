package bits

import (
	"bytes"
	"io"
	"testing"
)

func TestAccErrReader(t *testing.T) {
	input := []byte{0xff, 0x0f} // 1111 1111 0000 1111
	rd := bytes.NewReader(input)
	reader := NewAccErrReader(rd)

	cases := []struct {
		n    int
		want uint
	}{
		{2, 3},  // 11
		{3, 7},  // 111
		{5, 28}, // 11100
		{3, 1},  // 001
		{3, 7},  // 111
	}

	for _, tc := range cases {
		got := reader.Read(tc.n)

		if got != tc.want {
			t.Errorf("Read(%d)=%b, want=%b", tc.n, got, tc.want)
		}
	}
	err := reader.AccError()
	if err != nil {
		t.Errorf("Got accumulated error: %w", err)
	}
}

func TestBadAccErrReader(t *testing.T) {
	// Check that reading beyond EOF provides value = 0 after acc error
	input := []byte{0xff, 0x0f} // 1111 1111 0000 1111
	rd := bytes.NewReader(input)
	reader := NewAccErrReader(rd)

	cases := []struct {
		n    int
		want uint
		err  error
	}{
		{2, 3, nil},     // 11
		{3, 7, nil},     // 111
		{12, 0, io.EOF}, // 0 because of error
		{3, 0, io.EOF},  // 0 because of acc error
		{3, 0, io.EOF},  // 0 because of acc error
	}

	for _, tc := range cases {
		got := reader.Read(tc.n)

		if got != tc.want {
			t.Errorf("Read(%d)=%b, want=%b", tc.n, got, tc.want)
		}
	}
	err := reader.AccError()
	if err != io.EOF {
		t.Errorf("Wanted io.EOF but got %v", err)
	}
}
