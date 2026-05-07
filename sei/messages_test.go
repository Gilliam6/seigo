package sei

import (
	"bytes"
	"reflect"
	"testing"
)

func TestFillerPayload(t *testing.T) {
	src := &FillerPayload{Size: 5}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m.Payload, []byte{0xff, 0xff, 0xff, 0xff, 0xff}) {
		t.Fatalf("encode: % x", m.Payload)
	}
	got, err := ParseFillerPayload(m)
	if err != nil {
		t.Fatal(err)
	}
	if got.Size != 5 {
		t.Fatalf("size: %d", got.Size)
	}

	// Wrong byte triggers error.
	bad := Message{Type: PayloadTypeFillerPayload, Payload: []byte{0xff, 0x00}}
	if _, err := ParseFillerPayload(bad); err == nil {
		t.Fatal("expected error for non-0xFF byte")
	}
}

func TestUserDataUnregistered(t *testing.T) {
	src := &UserDataUnregistered{
		UUID: [16]byte{0xdc, 0x45, 0xe9, 0xbd, 0xe6, 0xd9, 0x48, 0xb7,
			0x96, 0x2c, 0xd8, 0x20, 0xd9, 0x23, 0xee, 0xef},
		Data: []byte("hello"),
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseUserDataUnregistered(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestUserDataUnregistered_TooShort(t *testing.T) {
	short := Message{Type: PayloadTypeUserDataUnregistered, Payload: bytes.Repeat([]byte{0}, 15)}
	if _, err := ParseUserDataUnregistered(short); err == nil {
		t.Fatal("expected error for truncated UUID")
	}
}

func TestUserDataRegisteredITUTT35_Single(t *testing.T) {
	src := &UserDataRegisteredITUTT35{CountryCode: 0xb5, Data: []byte{0x00, 0x31, 0x47, 0x41, 0x39, 0x34}}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if m.Payload[0] != 0xb5 {
		t.Fatalf("country code: % x", m.Payload)
	}
	got, err := ParseUserDataRegisteredITUTT35(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip: got %+v want %+v", got, src)
	}
}

func TestUserDataRegisteredITUTT35_Extended(t *testing.T) {
	src := &UserDataRegisteredITUTT35{CountryCode: 0xff, CountryCodeExtension: 0xa0, Data: []byte{0x12, 0x34}}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseUserDataRegisteredITUTT35(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip: got %+v want %+v", got, src)
	}
}

func TestRecoveryPoint(t *testing.T) {
	cases := []*RecoveryPoint{
		{RecoveryFrameCnt: 0, ExactMatchFlag: true, BrokenLinkFlag: false, ChangingSliceGroupIDC: 0},
		{RecoveryFrameCnt: 7, ExactMatchFlag: false, BrokenLinkFlag: true, ChangingSliceGroupIDC: 3},
		{RecoveryFrameCnt: 1024, ExactMatchFlag: true, BrokenLinkFlag: true, ChangingSliceGroupIDC: 2},
	}
	for _, src := range cases {
		m, err := src.Encode()
		if err != nil {
			t.Fatalf("encode %+v: %v", src, err)
		}
		got, err := ParseRecoveryPoint(m)
		if err != nil {
			t.Fatalf("parse %+v: %v", src, err)
		}
		if !reflect.DeepEqual(src, got) {
			t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
		}
	}
}

func TestRecoveryPoint_RejectsBadSliceGroupIDC(t *testing.T) {
	rp := &RecoveryPoint{ChangingSliceGroupIDC: 4}
	if _, err := rp.Encode(); err == nil {
		t.Fatal("expected error")
	}
}

func TestPanScanRect(t *testing.T) {
	src := &PanScanRect{
		ID:               1,
		CancelFlag:       false,
		Rectangles:       []PanScanRectangle{{1, -1, 2, -2}, {0, 0, 3, -3}},
		RepetitionPeriod: 5,
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParsePanScanRect(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestPanScanRect_Cancel(t *testing.T) {
	src := &PanScanRect{ID: 0, CancelFlag: true}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParsePanScanRect(m)
	if err != nil {
		t.Fatal(err)
	}
	if !got.CancelFlag || got.ID != 0 || len(got.Rectangles) != 0 || got.RepetitionPeriod != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestFramePackingArrangement(t *testing.T) {
	src := &FramePackingArrangement{
		ID:                       42,
		Type:                     3, // side-by-side
		QuincunxSamplingFlag:     false,
		ContentInterpretationType: 1,
		FieldViewsFlag:           false,
		CurrentFrameIsFrame0Flag: true,
		Frame0SelfContainedFlag:  true,
		Frame1SelfContainedFlag:  false,
		Frame0GridPositionX:      1,
		Frame0GridPositionY:      2,
		Frame1GridPositionX:      3,
		Frame1GridPositionY:      4,
		ReservedByte:             0,
		RepetitionPeriod:         0,
		ExtensionFlag:            false,
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseFramePackingArrangement(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestFramePackingArrangement_QuincunxSkipsGrid(t *testing.T) {
	src := &FramePackingArrangement{
		ID:                   1,
		Type:                 4,
		QuincunxSamplingFlag: true,
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseFramePackingArrangement(m)
	if err != nil {
		t.Fatal(err)
	}
	if got.Frame0GridPositionX != 0 || got.Frame1GridPositionY != 0 {
		t.Fatalf("grid positions should be 0 when quincunx flag is set: %+v", got)
	}
}

func TestFramePackingArrangement_Cancel(t *testing.T) {
	src := &FramePackingArrangement{ID: 99, CancelFlag: true, ExtensionFlag: false}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseFramePackingArrangement(m)
	if err != nil {
		t.Fatal(err)
	}
	if !got.CancelFlag || got.ID != 99 {
		t.Fatalf("got %+v", got)
	}
}

func TestDisplayOrientation(t *testing.T) {
	src := &DisplayOrientation{
		HorFlip:               true,
		VerFlip:               false,
		AnticlockwiseRotation: 0x4000, // 90 degrees
		RepetitionPeriod:      30,
		ExtensionFlag:         false,
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseDisplayOrientation(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestMasteringDisplayColourVolume(t *testing.T) {
	src := &MasteringDisplayColourVolume{
		DisplayPrimariesX:            [3]uint16{34000, 13250, 7500},
		DisplayPrimariesY:            [3]uint16{16000, 34500, 3000},
		WhitePointX:                  15635,
		WhitePointY:                  16450,
		MaxDisplayMasteringLuminance: 10000000,
		MinDisplayMasteringLuminance: 50,
	}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Payload) != 24 {
		t.Fatalf("payload size: %d", len(m.Payload))
	}
	got, err := ParseMasteringDisplayColourVolume(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestContentLightLevelInfo(t *testing.T) {
	src := &ContentLightLevelInfo{MaxContentLightLevel: 4000, MaxPicAverageLightLevel: 1000}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseContentLightLevelInfo(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip:\n got: %+v\nwant: %+v", got, src)
	}
}

func TestAlternativeTransferCharacteristics(t *testing.T) {
	src := &AlternativeTransferCharacteristics{PreferredTransferCharacteristics: 18}
	m, err := src.Encode()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseAlternativeTransferCharacteristics(m)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("round trip: got %+v want %+v", got, src)
	}
}

func TestPayloadType_String(t *testing.T) {
	if got := PayloadTypeRecoveryPoint.String(); got != "recovery_point" {
		t.Errorf("got %q", got)
	}
	if got := PayloadType(9999).String(); got != "PayloadType(9999)" {
		t.Errorf("got %q", got)
	}
}
