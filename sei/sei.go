// Package sei implements parsing and generation of H.264 Supplemental
// Enhancement Information (SEI) NAL units, as specified in Annex D of
// ITU-T Rec. H.264 / ISO/IEC 14496-10.
//
// The package operates at three layers:
//
//   - The raw NAL unit (with the 1-byte NAL unit header).
//   - The EBSP byte sequence (NAL unit body, including emulation prevention bytes).
//   - The RBSP byte sequence (NAL unit body with emulation prevention bytes removed).
//
// Use [ParseNALU] / [NALU.EncodeNALU] when you have a complete NAL unit, or
// the EBSP/RBSP variants when you have already stripped the header or the
// emulation prevention bytes from the byte stream.
//
// Each parsed [NALU] carries an ordered list of [Message] values. The raw
// payload bytes of each message are exposed verbatim, and typed encoders /
// decoders are provided for the most common payload types in this package.
package sei

import (
	"errors"
	"fmt"

	"github.com/Gilliam6/seigo/bits"
)

// NALUnitTypeSEI is the H.264 nal_unit_type value reserved for SEI NAL units.
const NALUnitTypeSEI = 6

// Errors returned by parsing routines.
var (
	ErrEmptyRBSP            = errors.New("sei: empty RBSP")
	ErrUnexpectedNALUType   = errors.New("sei: unexpected nal_unit_type")
	ErrInvalidTrailingBits  = errors.New("sei: invalid rbsp_trailing_bits")
	ErrPayloadSizeOverflow  = errors.New("sei: payload size exceeds remaining data")
	ErrPayloadTypeOverflow  = errors.New("sei: payload type exceeds uint32")
	ErrUnalignedAfterPayload = errors.New("sei: payload did not end on a byte boundary")
)

// Message represents one SEI message: a payload type and its raw payload
// bytes. The slice length equals the payloadSize field in the bitstream.
type Message struct {
	Type    PayloadType
	Payload []byte
}

// NALU is a parsed Supplemental Enhancement Information NAL unit. A SEI NAL
// unit always carries one or more messages in the order they appeared in the
// bitstream.
type NALU struct {
	// ForbiddenZeroBit is bit 0 of the NAL unit header. The H.264 spec
	// mandates 0 for any conforming bitstream.
	ForbiddenZeroBit uint8

	// NalRefIDC is the nal_ref_idc field of the NAL unit header. The H.264
	// specification (§7.4.1) requires nal_ref_idc to be 0 when nal_unit_type
	// is 6 (SEI), so [NALU.EncodeNALU] uses 0 unconditionally.
	NalRefIDC uint8

	// Messages are the SEI messages carried by this NAL unit.
	Messages []Message
}

// ParseRBSP parses sei_rbsp() (§7.3.2.3) from RBSP bytes — i.e. NAL unit body
// bytes with the emulation_prevention_three_byte already removed.
func ParseRBSP(rbsp []byte) (*NALU, error) {
	if len(rbsp) == 0 {
		return nil, ErrEmptyRBSP
	}
	n := &NALU{}
	r := bits.NewReader(rbsp)
	for r.HasMoreRBSPData() {
		msg, err := readMessage(r)
		if err != nil {
			return nil, err
		}
		if !r.ByteAligned() {
			return nil, ErrUnalignedAfterPayload
		}
		n.Messages = append(n.Messages, *msg)
	}
	if err := r.ReadRBSPTrailingBits(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTrailingBits, err)
	}
	return n, nil
}

// EncodeRBSP serializes the NALU's messages followed by rbsp_trailing_bits()
// into an RBSP byte sequence (no emulation prevention bytes).
func (n *NALU) EncodeRBSP() ([]byte, error) {
	w := bits.NewWriter()
	for i, msg := range n.Messages {
		if err := writeMessage(w, msg); err != nil {
			return nil, fmt.Errorf("sei: encoding message %d (%s): %w", i, msg.Type, err)
		}
		if !w.ByteAligned() {
			return nil, fmt.Errorf("sei: message %d produced unaligned output", i)
		}
	}
	w.WriteRBSPTrailingBits()
	return w.Bytes(), nil
}

// ParseEBSP parses a SEI NAL unit body from its EBSP byte sequence (i.e. the
// NAL unit body bytes that follow the 1-byte NAL unit header in the
// bitstream, with emulation_prevention_three_byte values still present).
func ParseEBSP(ebsp []byte) (*NALU, error) {
	return ParseRBSP(bits.RemoveEmulationPrevention(ebsp))
}

// EncodeEBSP returns the NAL unit body (EBSP) corresponding to this SEI
// message list. The NAL unit header is not included.
func (n *NALU) EncodeEBSP() ([]byte, error) {
	rbsp, err := n.EncodeRBSP()
	if err != nil {
		return nil, err
	}
	return bits.AddEmulationPrevention(rbsp), nil
}

// ParseNALU parses a complete SEI NAL unit, including its 1-byte header.
// The byte sequence must NOT include an Annex B start code prefix.
func ParseNALU(naluBytes []byte) (*NALU, error) {
	if len(naluBytes) < 1 {
		return nil, errors.New("sei: NAL unit too short")
	}
	hdr := naluBytes[0]
	nalUnitType := hdr & 0x1f
	if nalUnitType != NALUnitTypeSEI {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrUnexpectedNALUType, nalUnitType, NALUnitTypeSEI)
	}
	nalu, err := ParseEBSP(naluBytes[1:])
	if err != nil {
		return nil, err
	}
	nalu.ForbiddenZeroBit = (hdr >> 7) & 1
	nalu.NalRefIDC = (hdr >> 5) & 0x3
	return nalu, nil
}

// EncodeNALU returns the complete SEI NAL unit including its 1-byte header.
// No Annex B start code prefix is prepended. The H.264 spec mandates
// nal_ref_idc = 0 for SEI, so the encoded header is always 0x06.
func (n *NALU) EncodeNALU() ([]byte, error) {
	body, err := n.EncodeEBSP()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, 1+len(body))
	// forbidden_zero_bit=0, nal_ref_idc=0, nal_unit_type=6 → 0x06.
	out = append(out, NALUnitTypeSEI)
	out = append(out, body...)
	return out, nil
}

// readMessage parses one sei_message() (§7.3.2.3.1) from a byte-aligned
// reader. The returned payload slice has length equal to payloadSize.
func readMessage(r *bits.Reader) (*Message, error) {
	pt, err := readMultiByteValue(r, "payloadType")
	if err != nil {
		return nil, err
	}
	if pt > uint64(^uint32(0)) {
		return nil, ErrPayloadTypeOverflow
	}
	ps, err := readMultiByteValue(r, "payloadSize")
	if err != nil {
		return nil, err
	}
	if ps > uint64(r.BytesLeft()) {
		return nil, fmt.Errorf("%w: payloadSize=%d remaining=%d", ErrPayloadSizeOverflow, ps, r.BytesLeft())
	}
	payload, err := r.ReadBytes(int(ps))
	if err != nil {
		return nil, fmt.Errorf("sei: reading payload: %w", err)
	}
	return &Message{Type: PayloadType(pt), Payload: payload}, nil
}

// writeMessage serializes one sei_message() to a byte-aligned writer.
func writeMessage(w *bits.Writer, m Message) error {
	if !w.ByteAligned() {
		return errors.New("sei: writeMessage requires byte alignment")
	}
	writeMultiByteValue(w, uint64(m.Type))
	writeMultiByteValue(w, uint64(len(m.Payload)))
	return w.WriteBytes(m.Payload)
}

// readMultiByteValue decodes the 0xFF-extended unsigned integer encoding used
// for both payloadType and payloadSize: any number of 0xFF bytes followed by
// a final byte in [0, 0xFF]. The total value is the sum of all bytes read.
func readMultiByteValue(r *bits.Reader, what string) (uint64, error) {
	var v uint64
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("sei: reading %s byte: %w", what, err)
		}
		v += uint64(b)
		if b != 0xff {
			return v, nil
		}
	}
}

// writeMultiByteValue encodes an unsigned integer in the same 0xFF-extended
// format used by readMultiByteValue. The encoding is canonical: the longest
// possible run of 0xFF bytes is emitted, then a single terminator byte.
func writeMultiByteValue(w *bits.Writer, v uint64) {
	for v >= 0xff {
		_ = w.WriteByte(0xff)
		v -= 0xff
	}
	_ = w.WriteByte(byte(v))
}
