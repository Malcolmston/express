package forwarded

import (
	"reflect"
	"testing"
)

// Parity vectors transcribed verbatim from the upstream jshttp/forwarded test
// suite. Each vector's input header and expected output are taken from the real
// assertions in that file (the socket address in the upstream tests is always
// "127.0.0.1"), so these lock the Go port to the Node module's observed output.
//
// Upstream sources:
//   https://raw.githubusercontent.com/jshttp/forwarded/master/test/test.js
//   https://raw.githubusercontent.com/jshttp/forwarded/master/index.js

func TestParityForwarded(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		xff        string
		want       []string
	}{
		// it('should work with X-Forwarded-For header')
		{"no header", "127.0.0.1", "", []string{"127.0.0.1"}},
		// it('should include entries from X-Forwarded-For')
		{"includes entries", "127.0.0.1", "10.0.0.2, 10.0.0.1", []string{"127.0.0.1", "10.0.0.1", "10.0.0.2"}},
		// it('should skip blank entries')
		{"skip blank entries", "127.0.0.1", "10.0.0.2,, 10.0.0.1", []string{"127.0.0.1", "10.0.0.1", "10.0.0.2"}},
		// it('should trim leading OWS')
		{"trim OWS", "127.0.0.1", " 10.0.0.2 ,  , 10.0.0.1 ", []string{"127.0.0.1", "10.0.0.1", "10.0.0.2"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Forwarded(tc.remoteAddr, tc.xff)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Forwarded(%q, %q) = %v, want %v", tc.remoteAddr, tc.xff, got, tc.want)
			}
		})
	}
}

// describe('socket address') -> it('should begin with socket address')
func TestParitySocketAddressFirst(t *testing.T) {
	got := Forwarded("127.0.0.1", "")
	if len(got) == 0 || got[0] != "127.0.0.1" {
		t.Errorf("Forwarded socket address = %v, want first element %q", got, "127.0.0.1")
	}
}
