package mp4

import (
	"encoding/hex"
	"testing"
)

const (
	vpsHex = "40010c01ffff022000000300b0000003000003007b18b024"
	spsHex = "420101022000000300b0000003000003007ba0078200887db6718b92448053888892cf24a69272c9124922dc91aa48fca223ff000100016a02020201"
	ppsHex = "4401c0252f053240"
)

func TestHvcC(t *testing.T) {
	vpsNalu, err := hex.DecodeString(vpsHex)
	if err != nil {
		t.Error(err)
	}
	spsNalu, err := hex.DecodeString(spsHex)
	if err != nil {
		t.Error(err)
	}
	ppsNalu, err := hex.DecodeString(ppsHex)
	if err != nil {
		t.Error(err)
	}
	hvcC, err := CreateHvcC([][]byte{vpsNalu}, [][]byte{spsNalu}, [][]byte{ppsNalu}, true, true, true)
	if err != nil {
		t.Error(err)
	}
	boxDiffAfterEncodeAndDecode(t, hvcC)
}
