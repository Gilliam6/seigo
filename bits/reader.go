// Package bits provides bit-level readers and writers for H.264 bitstream syntax,
// including Exp-Golomb codes, RBSP trailing bits, and emulation-prevention-byte
// conversion between RBSP and EBSP.
package bits

import (
	"errors"
	"fmt"
	"io"
)

// ErrInvalidBitCount is returned when a caller asks for an invalid number of bits.
var ErrInvalidBitCount = errors.New("bits: invalid bit count")

// Reader is a sequential bit-level reader over an in-memory byte slice.
// It is the basic building block for parsing H.264 RBSP data such as
// Sequence Parameter Sets or SEI message payloads.
//
// Reader is not safe for concurrent use.
type Reader struct {
	data []byte
	pos  int // next bit to read, counted from the MSB of data[0]
}

// NewReader returns a new Reader positioned at the first bit of data.
// data is referenced, not copied.
func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

// BitsRead returns the number of bits already consumed.
func (r *Reader) BitsRead() int { return r.pos }

// BitsLeft returns the number of bits that remain to be read.
func (r *Reader) BitsLeft() int { return len(r.data)*8 - r.pos }

// BytesLeft returns the number of whole bytes that remain to be read,
// rounded down. Useful only when ByteAligned reports true.
func (r *Reader) BytesLeft() int { return r.BitsLeft() / 8 }

// ByteAligned reports whether the reader is currently at a byte boundary.
func (r *Reader) ByteAligned() bool { return r.pos%8 == 0 }

// Skip advances the reader by n bits.
func (r *Reader) Skip(n int) error {
	if n < 0 {
		return ErrInvalidBitCount
	}
	if r.pos+n > len(r.data)*8 {
		return io.ErrUnexpectedEOF
	}
	r.pos += n
	return nil
}

// AlignToByte rounds the read position up to the next byte boundary.
// Any skipped bits are discarded.
func (r *Reader) AlignToByte() {
	if r.pos%8 != 0 {
		r.pos += 8 - r.pos%8
	}
}

// PeekBit returns the next bit without advancing the read position.
func (r *Reader) PeekBit() (uint8, error) {
	if r.pos >= len(r.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}
	return (r.data[r.pos/8] >> (7 - uint(r.pos%8))) & 1, nil
}

// ReadBit reads a single bit (0 or 1).
func (r *Reader) ReadBit() (uint8, error) {
	b, err := r.PeekBit()
	if err != nil {
		return 0, err
	}
	r.pos++
	return b, nil
}

// ReadBool reads one bit and returns true if it equals 1.
// This corresponds to the u(1) descriptor in the H.264 spec.
func (r *Reader) ReadBool() (bool, error) {
	b, err := r.ReadBit()
	return b == 1, err
}

// ReadBits reads n bits (0 ≤ n ≤ 64) into the low-order bits of a uint64,
// big-endian. This corresponds to the u(n) and f(n) descriptors.
func (r *Reader) ReadBits(n int) (uint64, error) {
	if n < 0 || n > 64 {
		return 0, ErrInvalidBitCount
	}
	if r.pos+n > len(r.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}
	var v uint64
	for i := 0; i < n; i++ {
		bit := (r.data[r.pos/8] >> (7 - uint(r.pos%8))) & 1
		v = (v << 1) | uint64(bit)
		r.pos++
	}
	return v, nil
}

// ReadByte reads exactly 8 bits.
func (r *Reader) ReadByte() (byte, error) {
	v, err := r.ReadBits(8)
	return byte(v), err
}

// ReadBytes reads n whole bytes. The reader must be byte-aligned.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if !r.ByteAligned() {
		return nil, errors.New("bits: ReadBytes requires byte alignment")
	}
	if n < 0 {
		return nil, ErrInvalidBitCount
	}
	if r.BitsLeft() < n*8 {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]byte, n)
	copy(out, r.data[r.pos/8:r.pos/8+n])
	r.pos += n * 8
	return out, nil
}

// ReadUE reads an unsigned integer Exp-Golomb-coded value (descriptor ue(v))
// per ITU-T Rec. H.264 §9.1. Codewords longer than 32 leading zero bits are
// rejected because they cannot fit in a uint32.
func (r *Reader) ReadUE() (uint32, error) {
	leadingZeros := 0
	for {
		b, err := r.ReadBit()
		if err != nil {
			return 0, fmt.Errorf("bits: reading Exp-Golomb prefix: %w", err)
		}
		if b == 1 {
			break
		}
		leadingZeros++
		if leadingZeros > 32 {
			return 0, errors.New("bits: Exp-Golomb code too long")
		}
	}
	if leadingZeros == 0 {
		return 0, nil
	}
	suffix, err := r.ReadBits(leadingZeros)
	if err != nil {
		return 0, fmt.Errorf("bits: reading Exp-Golomb suffix: %w", err)
	}
	val := (uint64(1) << uint(leadingZeros)) - 1 + suffix
	if val > uint64(^uint32(0)) {
		return 0, errors.New("bits: Exp-Golomb value overflows uint32")
	}
	return uint32(val), nil
}

// ReadSE reads a signed integer Exp-Golomb-coded value (descriptor se(v))
// per ITU-T Rec. H.264 §9.1.1.
func (r *Reader) ReadSE() (int32, error) {
	code, err := r.ReadUE()
	if err != nil {
		return 0, err
	}
	if code&1 == 1 {
		return int32((code + 1) / 2), nil
	}
	return -int32(code / 2), nil
}

// HasMoreRBSPData implements the more_rbsp_data() function from §7.2 of the
// H.264 specification. It reports whether any data precedes rbsp_trailing_bits()
// at the current position. The trailing bits are defined as a single 1 bit
// followed by zero or more 0 bits up to a byte boundary, so this is equivalent
// to asking whether the last 1 bit in the buffer lies strictly after the
// current read position.
func (r *Reader) HasMoreRBSPData() bool {
	if r.pos >= len(r.data)*8 {
		return false
	}
	for last := len(r.data)*8 - 1; last >= r.pos; last-- {
		if (r.data[last/8]>>(7-uint(last%8)))&1 == 1 {
			return last > r.pos
		}
	}
	return false
}

// ReadRBSPTrailingBits validates that the remaining bits form a valid
// rbsp_trailing_bits() structure: a single 1 bit followed by 0 bits to the
// next byte boundary. The reader is advanced past those bits.
func (r *Reader) ReadRBSPTrailingBits() error {
	stopBit, err := r.ReadBit()
	if err != nil {
		return fmt.Errorf("bits: reading rbsp_stop_one_bit: %w", err)
	}
	if stopBit != 1 {
		return errors.New("bits: rbsp_stop_one_bit must be 1")
	}
	for !r.ByteAligned() {
		b, err := r.ReadBit()
		if err != nil {
			return fmt.Errorf("bits: reading rbsp_alignment_zero_bit: %w", err)
		}
		if b != 0 {
			return errors.New("bits: rbsp_alignment_zero_bit must be 0")
		}
	}
	return nil
}
