package bits

import (
	"bytes"
	"testing"
)

func TestWriter_WriteBits(t *testing.T) {
	w := NewWriter()
	_ = w.WriteBits(0b101, 3)
	_ = w.WriteBits(0b10010, 5)
	_ = w.WriteBits(0xC1, 8)
	got := w.Bytes()
	want := []byte{0xB2, 0xC1}
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x want % x", got, want)
	}
}

func TestWriter_WriteUE_RoundTrip(t *testing.T) {
	values := []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 100, 65535, 1<<20 - 1}
	for _, v := range values {
		w := NewWriter()
		w.WriteUE(v)
		w.AlignToByte()
		r := NewReader(w.Bytes())
		got, err := r.ReadUE()
		if err != nil {
			t.Errorf("ue(%d): %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("ue(%d) round trip: got %d", v, got)
		}
	}
}

func TestWriter_WriteSE_RoundTrip(t *testing.T) {
	values := []int32{0, 1, -1, 2, -2, 100, -100, 1 << 16, -(1 << 16)}
	for _, v := range values {
		w := NewWriter()
		w.WriteSE(v)
		w.AlignToByte()
		r := NewReader(w.Bytes())
		got, err := r.ReadSE()
		if err != nil {
			t.Errorf("se(%d): %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("se(%d) round trip: got %d", v, got)
		}
	}
}

func TestWriter_RBSPTrailingBits(t *testing.T) {
	w := NewWriter()
	_ = w.WriteByte(0x12)
	w.WriteRBSPTrailingBits()
	got := w.Bytes()
	want := []byte{0x12, 0x80}
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x want % x", got, want)
	}

	// Mid-byte trailing bits.
	w = NewWriter()
	_ = w.WriteBits(0b101, 3)
	w.WriteRBSPTrailingBits()
	if got, want := w.Bytes(), []byte{0b10110000}; !bytes.Equal(got, want) {
		t.Fatalf("mid-byte: got %08b want %08b", got, want)
	}
}

func TestWriter_SEIByteAlignment(t *testing.T) {
	// Already aligned: no-op.
	w := NewWriter()
	_ = w.WriteByte(0x12)
	w.WriteSEIByteAlignment()
	if got := w.Bytes(); !bytes.Equal(got, []byte{0x12}) {
		t.Fatalf("aligned: % x", got)
	}

	// Not aligned: write 1 bit + 0 bits to byte boundary.
	w = NewWriter()
	_ = w.WriteBits(0b101, 3)
	w.WriteSEIByteAlignment()
	if got := w.Bytes(); !bytes.Equal(got, []byte{0b10110000}) {
		t.Fatalf("unaligned: %08b", got)
	}
}
