package attachments

import "testing"

func TestDecodeDimensionsAllowedBoundaryAndOverflow(t *testing.T) {
	for _, tc := range []struct {
		w, h int
		want bool
	}{
		{1, 1, true}, {16384, 16384, true}, {16384, 16385, false},
		{MaxDecodedPixels, 1, true}, {MaxDecodedPixels + 1, 1, false},
		{0, 1, false}, {-1, 1, false}, {1, 0, false},
		{int(^uint(0) >> 1), int(^uint(0) >> 1), false},
	} {
		if got := DecodeDimensionsAllowed(tc.w, tc.h); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}
