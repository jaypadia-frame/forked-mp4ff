package mp4

import (
	"testing"
)

func TestEmsg(t *testing.T) {

	boxes := []Box{
		&EmsgBox{Version: 1,
			TimeScale:        90000,
			PresentationTime: 10000000,
			EventDuration:    90000,
			Id:               42,
			SchemeIdURI:      "https://aomedia.org/emsg/ID3",
			Value:            "relative",
		},
		&EmsgBox{Version: 0,
			TimeScale:             90000,
			PresentationTimeDelta: 45000,
			EventDuration:         90000,
			Id:                    42,
			SchemeIdURI:           "schid",
			Value:                 "special"},
	}

	for _, inBox := range boxes {
		boxDiffAfterEncodeAndDecode(t, inBox)
	}
}
