package sei

import (
	"bytes"
	"errors"
	"testing"
)

func TestNALU_RoundTrip(t *testing.T) {
	src := &NALU{
		Messages: []Message{
			{Type: PayloadTypeFillerPayload, Payload: []byte{0xff, 0xff, 0xff}},
			{Type: PayloadTypeUserDataUnregistered, Payload: append(
				[]byte{
					0xdc, 0x45, 0xe9, 0xbd, 0xe6, 0xd9, 0x48, 0xb7,
					0x96, 0x2c, 0xd8, 0x20, 0xd9, 0x23, 0xee, 0xef,
				},
				[]byte("hello world")...,
			)},
		},
	}
	naluBytes, err := src.EncodeNALU()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if naluBytes[0] != 0x06 {
		t.Fatalf("NAL header byte = %#x, want 0x06", naluBytes[0])
	}
	got, err := ParseNALU(naluBytes)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got.Messages) != len(src.Messages) {
		t.Fatalf("messages count: %d vs %d", len(got.Messages), len(src.Messages))
	}
	for i := range src.Messages {
		if got.Messages[i].Type != src.Messages[i].Type {
			t.Errorf("msg[%d].Type = %s, want %s", i, got.Messages[i].Type, src.Messages[i].Type)
		}
		if !bytes.Equal(got.Messages[i].Payload, src.Messages[i].Payload) {
			t.Errorf("msg[%d].Payload mismatch", i)
		}
	}
}

func TestNALU_RejectsWrongNALUType(t *testing.T) {
	// nal_unit_type = 7 (SPS), wrapped in a SEI-shaped body.
	naluBytes := []byte{0x07, 0x80}
	_, err := ParseNALU(naluBytes)
	if !errors.Is(err, ErrUnexpectedNALUType) {
		t.Fatalf("got %v, want ErrUnexpectedNALUType", err)
	}
}

func TestNALU_RejectsEmptyRBSP(t *testing.T) {
	if _, err := ParseRBSP(nil); !errors.Is(err, ErrEmptyRBSP) {
		t.Fatalf("got %v, want ErrEmptyRBSP", err)
	}
}

func TestNALU_RejectsTooShortNALU(t *testing.T) {
	if _, err := ParseNALU(nil); err == nil {
		t.Fatal("expected error for empty NALU")
	}
}

func TestNALU_PayloadSizeOverflow(t *testing.T) {
	// payloadType=4, payloadSize=255 (encoded as 0xFF 0x00), but only 5 payload bytes follow.
	rbsp := []byte{0x04, 0xFF, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x80}
	_, err := ParseRBSP(rbsp)
	if !errors.Is(err, ErrPayloadSizeOverflow) {
		t.Fatalf("got %v, want ErrPayloadSizeOverflow", err)
	}
}

func TestNALU_LargePayloadType(t *testing.T) {
	// payloadType encoded with two extension bytes (255 + 255 + 5 = 515).
	src := &NALU{Messages: []Message{{Type: 515, Payload: []byte{0xaa, 0xbb}}}}
	enc, err := src.EncodeNALU()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := ParseNALU(enc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got.Messages) != 1 || got.Messages[0].Type != 515 {
		t.Fatalf("got messages %+v", got.Messages)
	}
	if !bytes.Equal(got.Messages[0].Payload, []byte{0xaa, 0xbb}) {
		t.Fatalf("payload mismatch")
	}
}

func TestNALU_LargePayloadSize(t *testing.T) {
	// payloadSize = 300 (encoded as 0xFF 0x2D = 255 + 45).
	payload := bytes.Repeat([]byte{0x42}, 300)
	src := &NALU{Messages: []Message{{Type: PayloadTypeFillerPayload, Payload: bytes.Repeat([]byte{0xff}, 300)}, {Type: PayloadTypeUserDataUnregistered, Payload: payload}}}
	// Make UDU valid (must be ≥ 16 byte UUID): pad payload accordingly via the exact bytes we used.
	src.Messages[1].Payload = payload
	enc, err := src.EncodeNALU()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := ParseNALU(enc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !bytes.Equal(got.Messages[1].Payload, payload) {
		t.Fatalf("payload mismatch: len=%d", len(got.Messages[1].Payload))
	}
}

func TestNALU_EmulationPreventionPath(t *testing.T) {
	// Payload deliberately contains 0x00 0x00 0x00 to exercise the EBSP escaping.
	payload := []byte{
		0xdc, 0x45, 0xe9, 0xbd, 0xe6, 0xd9, 0x48, 0xb7,
		0x96, 0x2c, 0xd8, 0x20, 0xd9, 0x23, 0xee, 0xef,
		0x00, 0x00, 0x00, 0x01, 0x02,
	}
	src := &NALU{Messages: []Message{{Type: PayloadTypeUserDataUnregistered, Payload: payload}}}
	enc, err := src.EncodeNALU()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for i := 0; i+2 < len(enc); i++ {
		if enc[i] == 0 && enc[i+1] == 0 && enc[i+2] <= 0x01 {
			t.Fatalf("encoded NAL contains forbidden start code at %d: % x", i, enc)
		}
	}
	got, err := ParseNALU(enc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !bytes.Equal(got.Messages[0].Payload, payload) {
		t.Fatalf("payload mismatch:\n got: % x\nwant: % x", got.Messages[0].Payload, payload)
	}
}
