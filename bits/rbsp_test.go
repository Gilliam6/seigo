package bits

import (
	"bytes"
	"testing"
)

func TestEmulationPrevention_RoundTrip(t *testing.T) {
	cases := [][]byte{
		nil,
		{},
		{0x00, 0x00, 0x00},
		{0x00, 0x00, 0x01},
		{0x00, 0x00, 0x02},
		{0x00, 0x00, 0x03},
		{0x00, 0x00, 0x04},
		{0xFF, 0x00, 0x00, 0x00, 0x01, 0x00},
		{0x00, 0x00, 0x00, 0x00},
	}
	for _, in := range cases {
		ebsp := AddEmulationPrevention(in)
		// EBSP must never contain a start-code prefix.
		for i := 0; i+2 < len(ebsp); i++ {
			if ebsp[i] == 0 && ebsp[i+1] == 0 && ebsp[i+2] <= 0x03 && ebsp[i+2] != 0x03 {
				t.Errorf("RBSP=% x EBSP=% x: forbidden run at %d", in, ebsp, i)
			}
		}
		got := RemoveEmulationPrevention(ebsp)
		if !bytes.Equal(got, in) {
			t.Errorf("RBSP=% x → EBSP=% x → RBSP=% x", in, ebsp, got)
		}
	}
}

func TestEmulationPrevention_Insertion(t *testing.T) {
	cases := []struct {
		in, want []byte
	}{
		{[]byte{0x00, 0x00, 0x00}, []byte{0x00, 0x00, 0x03, 0x00}},
		{[]byte{0x00, 0x00, 0x01}, []byte{0x00, 0x00, 0x03, 0x01}},
		{[]byte{0x00, 0x00, 0x02}, []byte{0x00, 0x00, 0x03, 0x02}},
		{[]byte{0x00, 0x00, 0x03}, []byte{0x00, 0x00, 0x03, 0x03}},
		{[]byte{0x00, 0x00, 0x04}, []byte{0x00, 0x00, 0x04}},
		{[]byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00, 0x03, 0x00, 0x00}},
	}
	for _, c := range cases {
		got := AddEmulationPrevention(c.in)
		if !bytes.Equal(got, c.want) {
			t.Errorf("AddEmulationPrevention(% x): got % x want % x", c.in, got, c.want)
		}
	}
}

func TestRemoveEmulationPrevention_Standalone(t *testing.T) {
	// Standalone 0x03 (no preceding 00 00) must be passed through.
	in := []byte{0x12, 0x03, 0x34}
	got := RemoveEmulationPrevention(in)
	if !bytes.Equal(got, in) {
		t.Fatalf("got % x want % x", got, in)
	}
}
