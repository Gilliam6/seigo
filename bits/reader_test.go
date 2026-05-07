package bits

import (
	"errors"
	"io"
	"testing"
)

func TestReader_ReadBits(t *testing.T) {
	// 0b10110010 0b11000001 = 0xB2 0xC1
	r := NewReader([]byte{0xB2, 0xC1})
	if v, err := r.ReadBits(3); err != nil || v != 0b101 {
		t.Fatalf("first 3 bits: got %b err=%v", v, err)
	}
	if v, err := r.ReadBits(5); err != nil || v != 0b10010 {
		t.Fatalf("next 5 bits: got %b err=%v", v, err)
	}
	if v, err := r.ReadBits(8); err != nil || v != 0xC1 {
		t.Fatalf("last byte: got %x err=%v", v, err)
	}
	if _, err := r.ReadBit(); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestReader_ReadUE(t *testing.T) {
	// Codewords from §9.1 of the H.264 spec.
	cases := []struct {
		bits []byte
		bitN int
		want uint32
	}{
		{[]byte{0b10000000}, 1, 0},
		{[]byte{0b01000000}, 3, 1},
		{[]byte{0b01100000}, 3, 2},
		{[]byte{0b00100000}, 5, 3},
		{[]byte{0b00101000}, 5, 4},
		{[]byte{0b00110000}, 5, 5},
		{[]byte{0b00111000}, 5, 6},
		{[]byte{0b00010000}, 7, 7},
		{[]byte{0b00010010}, 7, 8},
	}
	for _, c := range cases {
		r := NewReader(c.bits)
		got, err := r.ReadUE()
		if err != nil {
			t.Errorf("ue(%08b): err=%v", c.bits[0], err)
			continue
		}
		if got != c.want {
			t.Errorf("ue(%08b): got %d want %d", c.bits[0], got, c.want)
		}
		if r.BitsRead() != c.bitN {
			t.Errorf("ue(%08b): consumed %d bits, want %d", c.bits[0], r.BitsRead(), c.bitN)
		}
	}
}

func TestReader_ReadSE(t *testing.T) {
	cases := []struct {
		ue   uint32
		want int32
	}{
		{0, 0},
		{1, 1},
		{2, -1},
		{3, 2},
		{4, -2},
		{5, 3},
		{6, -3},
	}
	for _, c := range cases {
		w := NewWriter()
		w.WriteUE(c.ue)
		w.AlignToByte()
		r := NewReader(w.Bytes())
		got, err := r.ReadSE()
		if err != nil {
			t.Errorf("se(%d): %v", c.ue, err)
			continue
		}
		if got != c.want {
			t.Errorf("se(ue=%d): got %d want %d", c.ue, got, c.want)
		}
	}
}

func TestReader_HasMoreRBSPData(t *testing.T) {
	// Single trailing byte 0x80 → no payload data, the only 1 bit is the stop bit.
	r := NewReader([]byte{0x80})
	if r.HasMoreRBSPData() {
		t.Fatalf("expected no more rbsp data for trailing-only buffer")
	}
	if err := r.ReadRBSPTrailingBits(); err != nil {
		t.Fatalf("trailing bits: %v", err)
	}

	// 0x12 0x80: data byte then trailing.
	r = NewReader([]byte{0x12, 0x80})
	if !r.HasMoreRBSPData() {
		t.Fatalf("expected data before trailing bits")
	}
	b, err := r.ReadByte()
	if err != nil || b != 0x12 {
		t.Fatalf("read data byte: %x %v", b, err)
	}
	if r.HasMoreRBSPData() {
		t.Fatalf("expected no more data after consuming payload")
	}
	if err := r.ReadRBSPTrailingBits(); err != nil {
		t.Fatalf("trailing bits after payload: %v", err)
	}
}

func TestReader_ReadRBSPTrailingBits_Invalid(t *testing.T) {
	// 0x00 has no stop bit.
	r := NewReader([]byte{0x00})
	if err := r.ReadRBSPTrailingBits(); err == nil {
		t.Fatalf("expected error for invalid trailing bits")
	}
}

func TestReader_AlignAndSkip(t *testing.T) {
	r := NewReader([]byte{0xAB, 0xCD})
	if _, err := r.ReadBits(3); err != nil {
		t.Fatal(err)
	}
	r.AlignToByte()
	if !r.ByteAligned() || r.BitsRead() != 8 {
		t.Fatalf("alignment failed: pos=%d aligned=%v", r.BitsRead(), r.ByteAligned())
	}
	if err := r.Skip(4); err != nil {
		t.Fatal(err)
	}
	if v, err := r.ReadBits(4); err != nil || v != 0xD {
		t.Fatalf("after skip: %x %v", v, err)
	}
}
