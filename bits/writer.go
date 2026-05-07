package bits

import (
	"errors"
)

// Writer is a sequential bit-level writer that buffers output in memory.
// It is the inverse of Reader and exposes the same descriptors used by the
// H.264 specification: u(n), ue(v), se(v) and friends.
//
// Writer is not safe for concurrent use.
type Writer struct {
	buf     []byte
	cur     byte
	bitPos  int // 0..7, position inside cur where the next bit will be written
}

// NewWriter returns an empty Writer.
func NewWriter() *Writer { return &Writer{} }

// BitsWritten returns the number of bits written so far.
func (w *Writer) BitsWritten() int { return len(w.buf)*8 + w.bitPos }

// ByteAligned reports whether the writer is currently at a byte boundary.
func (w *Writer) ByteAligned() bool { return w.bitPos == 0 }

// WriteBit writes the least-significant bit of b.
func (w *Writer) WriteBit(b uint8) {
	w.cur |= (b & 1) << uint(7-w.bitPos)
	w.bitPos++
	if w.bitPos == 8 {
		w.buf = append(w.buf, w.cur)
		w.cur = 0
		w.bitPos = 0
	}
}

// WriteBool writes a single 1 or 0 bit.
func (w *Writer) WriteBool(v bool) {
	if v {
		w.WriteBit(1)
	} else {
		w.WriteBit(0)
	}
}

// WriteBits writes the n least-significant bits of v, big-endian.
func (w *Writer) WriteBits(v uint64, n int) error {
	if n < 0 || n > 64 {
		return ErrInvalidBitCount
	}
	for i := n - 1; i >= 0; i-- {
		w.WriteBit(uint8((v >> uint(i)) & 1))
	}
	return nil
}

// WriteByte writes 8 bits. It always returns nil and exists primarily so the
// writer satisfies io.ByteWriter.
func (w *Writer) WriteByte(b byte) error {
	return w.WriteBits(uint64(b), 8)
}

// WriteBytes writes len(p) bytes. The writer must be byte-aligned.
func (w *Writer) WriteBytes(p []byte) error {
	if !w.ByteAligned() {
		return errors.New("bits: WriteBytes requires byte alignment")
	}
	w.buf = append(w.buf, p...)
	return nil
}

// AlignToByte rounds the write position up to the next byte boundary, padding
// with zero bits. If already aligned, this is a no-op.
func (w *Writer) AlignToByte() {
	if w.bitPos != 0 {
		w.buf = append(w.buf, w.cur)
		w.cur = 0
		w.bitPos = 0
	}
}

// WriteUE writes an unsigned integer using the Exp-Golomb coding defined in
// §9.1 of the H.264 specification (descriptor ue(v)).
func (w *Writer) WriteUE(v uint32) {
	val := uint64(v) + 1
	leadingZeros := 0
	for tmp := val >> 1; tmp != 0; tmp >>= 1 {
		leadingZeros++
	}
	_ = w.WriteBits(0, leadingZeros)
	_ = w.WriteBits(val, leadingZeros+1)
}

// WriteSE writes a signed integer using the Exp-Golomb coding defined in
// §9.1.1 of the H.264 specification (descriptor se(v)).
func (w *Writer) WriteSE(v int32) {
	var u uint32
	if v <= 0 {
		u = uint32(-int64(v)) * 2
	} else {
		u = uint32(v)*2 - 1
	}
	w.WriteUE(u)
}

// WriteRBSPTrailingBits writes rbsp_trailing_bits(): a single 1 bit followed
// by zero or more 0 bits up to the next byte boundary.
func (w *Writer) WriteRBSPTrailingBits() {
	w.WriteBit(1)
	w.AlignToByte()
}

// WriteSEIByteAlignment writes the byte_alignment trailer used inside
// sei_payload() (§D.2): a 1 bit followed by 0 bits up to a byte boundary,
// only when the writer is not already byte-aligned.
func (w *Writer) WriteSEIByteAlignment() {
	if w.ByteAligned() {
		return
	}
	w.WriteBit(1)
	for !w.ByteAligned() {
		w.WriteBit(0)
	}
}

// Bytes returns a copy of all bytes written so far. If the writer is not
// byte-aligned, the partially filled last byte is included as well, padded
// with trailing zero bits.
func (w *Writer) Bytes() []byte {
	if w.bitPos == 0 {
		out := make([]byte, len(w.buf))
		copy(out, w.buf)
		return out
	}
	out := make([]byte, len(w.buf)+1)
	copy(out, w.buf)
	out[len(w.buf)] = w.cur
	return out
}
