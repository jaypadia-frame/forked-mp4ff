package avc

import (
	"testing"

	"github.com/go-test/deep"
)

func TestGetNaluTypes(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		wanted []NaluType
	}{
		{
			"IDR",
			[]byte{0, 0, 0, 2, 5},
			[]NaluType{NALU_IDR},
		},
		{
			"AUD and SPS",
			[]byte{0, 0, 0, 2, 9, 2, 0, 0, 0, 3, 7, 5, 4},
			[]NaluType{NALU_AUD, NALU_SPS},
		},
	}

	for _, tc := range testCases {
		got := FindNaluTypes(tc.input)
		if diff := deep.Equal(got, tc.wanted); diff != nil {
			t.Errorf("%s: %v", tc.name, diff)
		}
	}
}
