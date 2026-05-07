package sei

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/Gilliam6/seigo/bits"
)

// All structs in this file follow the same convention:
//
//   - The Encode method produces a [Message] suitable for inclusion in a
//     [NALU]. It applies whatever byte alignment the spec requires for the
//     specific payload.
//   - A Parse* function reverses the Encode for the matching payload type
//     and validates the message tag.
//
// Where the H.264 syntax tables describe payloads with conditional fields
// gated on a flag (e.g. a *_cancel_flag), the corresponding Go struct keeps
// that flag and treats the dependent fields as zero-valued when the flag
// indicates "cancel".

// -----------------------------------------------------------------------------
// filler_payload (payload type 3, §D.1.4)
// -----------------------------------------------------------------------------

// FillerPayload is a SEI message of type filler_payload. Its serialized form
// is simply a run of ff_byte values.
type FillerPayload struct {
	// Size is the number of 0xFF bytes the message carries.
	Size int
}

// Encode produces a Message representation of this filler payload.
func (f *FillerPayload) Encode() (Message, error) {
	if f.Size < 0 {
		return Message{}, errors.New("sei: filler_payload size must be non-negative")
	}
	payload := make([]byte, f.Size)
	for i := range payload {
		payload[i] = 0xff
	}
	return Message{Type: PayloadTypeFillerPayload, Payload: payload}, nil
}

// ParseFillerPayload validates that m carries a well-formed filler_payload
// and returns its decoded representation.
func ParseFillerPayload(m Message) (*FillerPayload, error) {
	if m.Type != PayloadTypeFillerPayload {
		return nil, fmt.Errorf("sei: not filler_payload: %s", m.Type)
	}
	for i, b := range m.Payload {
		if b != 0xff {
			return nil, fmt.Errorf("sei: filler_payload byte %d is 0x%02x, want 0xFF", i, b)
		}
	}
	return &FillerPayload{Size: len(m.Payload)}, nil
}

// -----------------------------------------------------------------------------
// user_data_unregistered (payload type 5, §D.1.6)
// -----------------------------------------------------------------------------

// UserDataUnregistered is a SEI message of type user_data_unregistered. It
// carries arbitrary application-specific data tagged with a 16-byte UUID
// (uuid_iso_iec_11578) chosen by the producer.
type UserDataUnregistered struct {
	UUID [16]byte
	Data []byte
}

// Encode produces a Message representation of the user_data_unregistered.
func (u *UserDataUnregistered) Encode() (Message, error) {
	payload := make([]byte, 0, 16+len(u.Data))
	payload = append(payload, u.UUID[:]...)
	payload = append(payload, u.Data...)
	return Message{Type: PayloadTypeUserDataUnregistered, Payload: payload}, nil
}

// ParseUserDataUnregistered decodes m as user_data_unregistered.
func ParseUserDataUnregistered(m Message) (*UserDataUnregistered, error) {
	if m.Type != PayloadTypeUserDataUnregistered {
		return nil, fmt.Errorf("sei: not user_data_unregistered: %s", m.Type)
	}
	if len(m.Payload) < 16 {
		return nil, errors.New("sei: user_data_unregistered missing UUID")
	}
	out := &UserDataUnregistered{}
	copy(out.UUID[:], m.Payload[:16])
	out.Data = append([]byte(nil), m.Payload[16:]...)
	return out, nil
}

// -----------------------------------------------------------------------------
// user_data_registered_itu_t_t35 (payload type 4, §D.1.5)
// -----------------------------------------------------------------------------

// UserDataRegisteredITUTT35 is a SEI message of type
// user_data_registered_itu_t_t35. The country code follows ITU-T T.35; if it
// is 0xFF, an additional CountryCodeExtension byte is present.
type UserDataRegisteredITUTT35 struct {
	CountryCode          uint8
	CountryCodeExtension uint8
	Data                 []byte
}

// Encode produces a Message representation of the registered user data.
func (u *UserDataRegisteredITUTT35) Encode() (Message, error) {
	payload := make([]byte, 0, 2+len(u.Data))
	payload = append(payload, u.CountryCode)
	if u.CountryCode == 0xff {
		payload = append(payload, u.CountryCodeExtension)
	}
	payload = append(payload, u.Data...)
	return Message{Type: PayloadTypeUserDataRegisteredITUTT35, Payload: payload}, nil
}

// ParseUserDataRegisteredITUTT35 decodes m as user_data_registered_itu_t_t35.
func ParseUserDataRegisteredITUTT35(m Message) (*UserDataRegisteredITUTT35, error) {
	if m.Type != PayloadTypeUserDataRegisteredITUTT35 {
		return nil, fmt.Errorf("sei: not user_data_registered_itu_t_t35: %s", m.Type)
	}
	if len(m.Payload) < 1 {
		return nil, errors.New("sei: user_data_registered_itu_t_t35 missing country code")
	}
	out := &UserDataRegisteredITUTT35{CountryCode: m.Payload[0]}
	offset := 1
	if out.CountryCode == 0xff {
		if len(m.Payload) < 2 {
			return nil, errors.New("sei: user_data_registered_itu_t_t35 missing country_code_extension_byte")
		}
		out.CountryCodeExtension = m.Payload[1]
		offset = 2
	}
	out.Data = append([]byte(nil), m.Payload[offset:]...)
	return out, nil
}

// -----------------------------------------------------------------------------
// recovery_point (payload type 6, §D.1.7)
// -----------------------------------------------------------------------------

// RecoveryPoint signals a recovery point for random access.
type RecoveryPoint struct {
	RecoveryFrameCnt        uint32
	ExactMatchFlag          bool
	BrokenLinkFlag          bool
	ChangingSliceGroupIDC   uint8 // valid range 0..3
}

// Encode produces a Message representation of the recovery_point payload.
func (rp *RecoveryPoint) Encode() (Message, error) {
	if rp.ChangingSliceGroupIDC > 3 {
		return Message{}, errors.New("sei: changing_slice_group_idc must be 0..3")
	}
	w := bits.NewWriter()
	w.WriteUE(rp.RecoveryFrameCnt)
	w.WriteBool(rp.ExactMatchFlag)
	w.WriteBool(rp.BrokenLinkFlag)
	_ = w.WriteBits(uint64(rp.ChangingSliceGroupIDC), 2)
	w.WriteSEIByteAlignment()
	return Message{Type: PayloadTypeRecoveryPoint, Payload: w.Bytes()}, nil
}

// ParseRecoveryPoint decodes m as recovery_point.
func ParseRecoveryPoint(m Message) (*RecoveryPoint, error) {
	if m.Type != PayloadTypeRecoveryPoint {
		return nil, fmt.Errorf("sei: not recovery_point: %s", m.Type)
	}
	r := bits.NewReader(m.Payload)
	rp := &RecoveryPoint{}
	var err error
	if rp.RecoveryFrameCnt, err = r.ReadUE(); err != nil {
		return nil, fmt.Errorf("recovery_frame_cnt: %w", err)
	}
	if rp.ExactMatchFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("exact_match_flag: %w", err)
	}
	if rp.BrokenLinkFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("broken_link_flag: %w", err)
	}
	v, err := r.ReadBits(2)
	if err != nil {
		return nil, fmt.Errorf("changing_slice_group_idc: %w", err)
	}
	rp.ChangingSliceGroupIDC = uint8(v)
	return rp, nil
}

// -----------------------------------------------------------------------------
// pan_scan_rect (payload type 2, §D.1.3)
// -----------------------------------------------------------------------------

// PanScanRectangle describes a single pan-scan rectangle.
type PanScanRectangle struct {
	LeftOffset   int32
	RightOffset  int32
	TopOffset    int32
	BottomOffset int32
}

// PanScanRect signals one or more pan-scan display rectangles.
type PanScanRect struct {
	ID               uint32
	CancelFlag       bool
	Rectangles       []PanScanRectangle
	RepetitionPeriod uint32
}

// Encode produces a Message representation of the pan_scan_rect payload.
func (p *PanScanRect) Encode() (Message, error) {
	w := bits.NewWriter()
	w.WriteUE(p.ID)
	w.WriteBool(p.CancelFlag)
	if !p.CancelFlag {
		if len(p.Rectangles) == 0 {
			return Message{}, errors.New("sei: pan_scan_rect requires at least one rectangle when not cancelled")
		}
		w.WriteUE(uint32(len(p.Rectangles) - 1))
		for _, rect := range p.Rectangles {
			w.WriteSE(rect.LeftOffset)
			w.WriteSE(rect.RightOffset)
			w.WriteSE(rect.TopOffset)
			w.WriteSE(rect.BottomOffset)
		}
		w.WriteUE(p.RepetitionPeriod)
	}
	w.WriteSEIByteAlignment()
	return Message{Type: PayloadTypePanScanRect, Payload: w.Bytes()}, nil
}

// ParsePanScanRect decodes m as pan_scan_rect.
func ParsePanScanRect(m Message) (*PanScanRect, error) {
	if m.Type != PayloadTypePanScanRect {
		return nil, fmt.Errorf("sei: not pan_scan_rect: %s", m.Type)
	}
	r := bits.NewReader(m.Payload)
	p := &PanScanRect{}
	var err error
	if p.ID, err = r.ReadUE(); err != nil {
		return nil, fmt.Errorf("pan_scan_rect_id: %w", err)
	}
	if p.CancelFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("pan_scan_rect_cancel_flag: %w", err)
	}
	if !p.CancelFlag {
		cnt, err := r.ReadUE()
		if err != nil {
			return nil, fmt.Errorf("pan_scan_cnt_minus1: %w", err)
		}
		if cnt > 2 {
			// §D.2.3: pan_scan_cnt_minus1 must be in [0, 2].
			return nil, fmt.Errorf("sei: pan_scan_cnt_minus1=%d out of range", cnt)
		}
		p.Rectangles = make([]PanScanRectangle, cnt+1)
		for i := range p.Rectangles {
			if p.Rectangles[i].LeftOffset, err = r.ReadSE(); err != nil {
				return nil, fmt.Errorf("pan_scan_rect_left_offset[%d]: %w", i, err)
			}
			if p.Rectangles[i].RightOffset, err = r.ReadSE(); err != nil {
				return nil, fmt.Errorf("pan_scan_rect_right_offset[%d]: %w", i, err)
			}
			if p.Rectangles[i].TopOffset, err = r.ReadSE(); err != nil {
				return nil, fmt.Errorf("pan_scan_rect_top_offset[%d]: %w", i, err)
			}
			if p.Rectangles[i].BottomOffset, err = r.ReadSE(); err != nil {
				return nil, fmt.Errorf("pan_scan_rect_bottom_offset[%d]: %w", i, err)
			}
		}
		if p.RepetitionPeriod, err = r.ReadUE(); err != nil {
			return nil, fmt.Errorf("pan_scan_rect_repetition_period: %w", err)
		}
	}
	return p, nil
}

// -----------------------------------------------------------------------------
// frame_packing_arrangement (payload type 45, §D.1.27)
// -----------------------------------------------------------------------------

// FramePackingArrangement signals stereoscopic frame-packing layout.
type FramePackingArrangement struct {
	ID                       uint32
	CancelFlag               bool
	Type                     uint8 // 0..7, only valid if !CancelFlag
	QuincunxSamplingFlag     bool
	ContentInterpretationType uint8 // 0..63
	SpatialFlippingFlag      bool
	Frame0FlippedFlag        bool
	FieldViewsFlag           bool
	CurrentFrameIsFrame0Flag bool
	Frame0SelfContainedFlag  bool
	Frame1SelfContainedFlag  bool
	Frame0GridPositionX      uint8 // 0..15
	Frame0GridPositionY      uint8
	Frame1GridPositionX      uint8
	Frame1GridPositionY      uint8
	ReservedByte             uint8
	RepetitionPeriod         uint32
	ExtensionFlag            bool
}

// Encode produces a Message representation of frame_packing_arrangement.
func (f *FramePackingArrangement) Encode() (Message, error) {
	w := bits.NewWriter()
	w.WriteUE(f.ID)
	w.WriteBool(f.CancelFlag)
	if !f.CancelFlag {
		if f.Type > 0x7f {
			return Message{}, errors.New("sei: frame_packing_arrangement_type must be 0..127")
		}
		_ = w.WriteBits(uint64(f.Type), 7)
		w.WriteBool(f.QuincunxSamplingFlag)
		if f.ContentInterpretationType > 0x3f {
			return Message{}, errors.New("sei: content_interpretation_type must be 0..63")
		}
		_ = w.WriteBits(uint64(f.ContentInterpretationType), 6)
		w.WriteBool(f.SpatialFlippingFlag)
		w.WriteBool(f.Frame0FlippedFlag)
		w.WriteBool(f.FieldViewsFlag)
		w.WriteBool(f.CurrentFrameIsFrame0Flag)
		w.WriteBool(f.Frame0SelfContainedFlag)
		w.WriteBool(f.Frame1SelfContainedFlag)
		if !f.QuincunxSamplingFlag && f.Type != 5 {
			if f.Frame0GridPositionX > 15 || f.Frame0GridPositionY > 15 ||
				f.Frame1GridPositionX > 15 || f.Frame1GridPositionY > 15 {
				return Message{}, errors.New("sei: frame*_grid_position_* must be 0..15")
			}
			_ = w.WriteBits(uint64(f.Frame0GridPositionX), 4)
			_ = w.WriteBits(uint64(f.Frame0GridPositionY), 4)
			_ = w.WriteBits(uint64(f.Frame1GridPositionX), 4)
			_ = w.WriteBits(uint64(f.Frame1GridPositionY), 4)
		}
		_ = w.WriteByte(f.ReservedByte)
		w.WriteUE(f.RepetitionPeriod)
	}
	w.WriteBool(f.ExtensionFlag)
	w.WriteSEIByteAlignment()
	return Message{Type: PayloadTypeFramePackingArrangement, Payload: w.Bytes()}, nil
}

// ParseFramePackingArrangement decodes m as frame_packing_arrangement.
func ParseFramePackingArrangement(m Message) (*FramePackingArrangement, error) {
	if m.Type != PayloadTypeFramePackingArrangement {
		return nil, fmt.Errorf("sei: not frame_packing_arrangement: %s", m.Type)
	}
	r := bits.NewReader(m.Payload)
	f := &FramePackingArrangement{}
	var err error
	if f.ID, err = r.ReadUE(); err != nil {
		return nil, fmt.Errorf("frame_packing_arrangement_id: %w", err)
	}
	if f.CancelFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("frame_packing_arrangement_cancel_flag: %w", err)
	}
	if !f.CancelFlag {
		v, err := r.ReadBits(7)
		if err != nil {
			return nil, fmt.Errorf("frame_packing_arrangement_type: %w", err)
		}
		f.Type = uint8(v)
		if f.QuincunxSamplingFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("quincunx_sampling_flag: %w", err)
		}
		v, err = r.ReadBits(6)
		if err != nil {
			return nil, fmt.Errorf("content_interpretation_type: %w", err)
		}
		f.ContentInterpretationType = uint8(v)
		if f.SpatialFlippingFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("spatial_flipping_flag: %w", err)
		}
		if f.Frame0FlippedFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("frame0_flipped_flag: %w", err)
		}
		if f.FieldViewsFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("field_views_flag: %w", err)
		}
		if f.CurrentFrameIsFrame0Flag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("current_frame_is_frame0_flag: %w", err)
		}
		if f.Frame0SelfContainedFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("frame0_self_contained_flag: %w", err)
		}
		if f.Frame1SelfContainedFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("frame1_self_contained_flag: %w", err)
		}
		if !f.QuincunxSamplingFlag && f.Type != 5 {
			if v, err = r.ReadBits(4); err != nil {
				return nil, fmt.Errorf("frame0_grid_position_x: %w", err)
			}
			f.Frame0GridPositionX = uint8(v)
			if v, err = r.ReadBits(4); err != nil {
				return nil, fmt.Errorf("frame0_grid_position_y: %w", err)
			}
			f.Frame0GridPositionY = uint8(v)
			if v, err = r.ReadBits(4); err != nil {
				return nil, fmt.Errorf("frame1_grid_position_x: %w", err)
			}
			f.Frame1GridPositionX = uint8(v)
			if v, err = r.ReadBits(4); err != nil {
				return nil, fmt.Errorf("frame1_grid_position_y: %w", err)
			}
			f.Frame1GridPositionY = uint8(v)
		}
		b, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("frame_packing_arrangement_reserved_byte: %w", err)
		}
		f.ReservedByte = b
		if f.RepetitionPeriod, err = r.ReadUE(); err != nil {
			return nil, fmt.Errorf("frame_packing_arrangement_repetition_period: %w", err)
		}
	}
	if f.ExtensionFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("frame_packing_arrangement_extension_flag: %w", err)
	}
	return f, nil
}

// -----------------------------------------------------------------------------
// display_orientation (payload type 47, §D.1.29)
// -----------------------------------------------------------------------------

// DisplayOrientation signals horizontal/vertical flip and rotation hints.
type DisplayOrientation struct {
	CancelFlag            bool
	HorFlip               bool
	VerFlip               bool
	AnticlockwiseRotation uint16 // u(16), units of 1/65536 of a full rotation
	RepetitionPeriod      uint32
	ExtensionFlag         bool
}

// Encode produces a Message representation of display_orientation.
func (d *DisplayOrientation) Encode() (Message, error) {
	w := bits.NewWriter()
	w.WriteBool(d.CancelFlag)
	if !d.CancelFlag {
		w.WriteBool(d.HorFlip)
		w.WriteBool(d.VerFlip)
		_ = w.WriteBits(uint64(d.AnticlockwiseRotation), 16)
		w.WriteUE(d.RepetitionPeriod)
		w.WriteBool(d.ExtensionFlag)
	}
	w.WriteSEIByteAlignment()
	return Message{Type: PayloadTypeDisplayOrientation, Payload: w.Bytes()}, nil
}

// ParseDisplayOrientation decodes m as display_orientation.
func ParseDisplayOrientation(m Message) (*DisplayOrientation, error) {
	if m.Type != PayloadTypeDisplayOrientation {
		return nil, fmt.Errorf("sei: not display_orientation: %s", m.Type)
	}
	r := bits.NewReader(m.Payload)
	d := &DisplayOrientation{}
	var err error
	if d.CancelFlag, err = r.ReadBool(); err != nil {
		return nil, fmt.Errorf("display_orientation_cancel_flag: %w", err)
	}
	if !d.CancelFlag {
		if d.HorFlip, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("hor_flip: %w", err)
		}
		if d.VerFlip, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("ver_flip: %w", err)
		}
		v, err := r.ReadBits(16)
		if err != nil {
			return nil, fmt.Errorf("anticlockwise_rotation: %w", err)
		}
		d.AnticlockwiseRotation = uint16(v)
		if d.RepetitionPeriod, err = r.ReadUE(); err != nil {
			return nil, fmt.Errorf("display_orientation_repetition_period: %w", err)
		}
		if d.ExtensionFlag, err = r.ReadBool(); err != nil {
			return nil, fmt.Errorf("display_orientation_extension_flag: %w", err)
		}
	}
	return d, nil
}

// -----------------------------------------------------------------------------
// mastering_display_colour_volume (payload type 137)
// -----------------------------------------------------------------------------

// MasteringDisplayColourVolume conveys mastering display characteristics for
// HDR signaling. The payload is a fixed 24 bytes laid out big-endian.
type MasteringDisplayColourVolume struct {
	DisplayPrimariesX            [3]uint16
	DisplayPrimariesY            [3]uint16
	WhitePointX                  uint16
	WhitePointY                  uint16
	MaxDisplayMasteringLuminance uint32
	MinDisplayMasteringLuminance uint32
}

// Encode produces a Message representation of mastering_display_colour_volume.
func (m *MasteringDisplayColourVolume) Encode() (Message, error) {
	payload := make([]byte, 24)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint16(payload[i*4:], m.DisplayPrimariesX[i])
		binary.BigEndian.PutUint16(payload[i*4+2:], m.DisplayPrimariesY[i])
	}
	binary.BigEndian.PutUint16(payload[12:], m.WhitePointX)
	binary.BigEndian.PutUint16(payload[14:], m.WhitePointY)
	binary.BigEndian.PutUint32(payload[16:], m.MaxDisplayMasteringLuminance)
	binary.BigEndian.PutUint32(payload[20:], m.MinDisplayMasteringLuminance)
	return Message{Type: PayloadTypeMasteringDisplayColourVolume, Payload: payload}, nil
}

// ParseMasteringDisplayColourVolume decodes m as mastering_display_colour_volume.
func ParseMasteringDisplayColourVolume(msg Message) (*MasteringDisplayColourVolume, error) {
	if msg.Type != PayloadTypeMasteringDisplayColourVolume {
		return nil, fmt.Errorf("sei: not mastering_display_colour_volume: %s", msg.Type)
	}
	if len(msg.Payload) != 24 {
		return nil, fmt.Errorf("sei: mastering_display_colour_volume must be 24 bytes, got %d", len(msg.Payload))
	}
	out := &MasteringDisplayColourVolume{}
	for i := 0; i < 3; i++ {
		out.DisplayPrimariesX[i] = binary.BigEndian.Uint16(msg.Payload[i*4:])
		out.DisplayPrimariesY[i] = binary.BigEndian.Uint16(msg.Payload[i*4+2:])
	}
	out.WhitePointX = binary.BigEndian.Uint16(msg.Payload[12:])
	out.WhitePointY = binary.BigEndian.Uint16(msg.Payload[14:])
	out.MaxDisplayMasteringLuminance = binary.BigEndian.Uint32(msg.Payload[16:])
	out.MinDisplayMasteringLuminance = binary.BigEndian.Uint32(msg.Payload[20:])
	return out, nil
}

// -----------------------------------------------------------------------------
// content_light_level_info (payload type 144)
// -----------------------------------------------------------------------------

// ContentLightLevelInfo conveys MaxCLL and MaxFALL HDR metadata. Payload is
// exactly 4 bytes (two big-endian uint16 values).
type ContentLightLevelInfo struct {
	MaxContentLightLevel    uint16
	MaxPicAverageLightLevel uint16
}

// Encode produces a Message representation of content_light_level_info.
func (c *ContentLightLevelInfo) Encode() (Message, error) {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:], c.MaxContentLightLevel)
	binary.BigEndian.PutUint16(payload[2:], c.MaxPicAverageLightLevel)
	return Message{Type: PayloadTypeContentLightLevelInfo, Payload: payload}, nil
}

// ParseContentLightLevelInfo decodes m as content_light_level_info.
func ParseContentLightLevelInfo(m Message) (*ContentLightLevelInfo, error) {
	if m.Type != PayloadTypeContentLightLevelInfo {
		return nil, fmt.Errorf("sei: not content_light_level_info: %s", m.Type)
	}
	if len(m.Payload) != 4 {
		return nil, fmt.Errorf("sei: content_light_level_info must be 4 bytes, got %d", len(m.Payload))
	}
	return &ContentLightLevelInfo{
		MaxContentLightLevel:    binary.BigEndian.Uint16(m.Payload[0:]),
		MaxPicAverageLightLevel: binary.BigEndian.Uint16(m.Payload[2:]),
	}, nil
}

// -----------------------------------------------------------------------------
// alternative_transfer_characteristics (payload type 147)
// -----------------------------------------------------------------------------

// AlternativeTransferCharacteristics overrides VUI transfer_characteristics
// for HDR-aware decoders. Payload is a single byte.
type AlternativeTransferCharacteristics struct {
	PreferredTransferCharacteristics uint8
}

// Encode produces a Message representation of alternative_transfer_characteristics.
func (a *AlternativeTransferCharacteristics) Encode() (Message, error) {
	return Message{
		Type:    PayloadTypeAlternativeTransferCharacteristics,
		Payload: []byte{a.PreferredTransferCharacteristics},
	}, nil
}

// ParseAlternativeTransferCharacteristics decodes m as alternative_transfer_characteristics.
func ParseAlternativeTransferCharacteristics(m Message) (*AlternativeTransferCharacteristics, error) {
	if m.Type != PayloadTypeAlternativeTransferCharacteristics {
		return nil, fmt.Errorf("sei: not alternative_transfer_characteristics: %s", m.Type)
	}
	if len(m.Payload) != 1 {
		return nil, fmt.Errorf("sei: alternative_transfer_characteristics must be 1 byte, got %d", len(m.Payload))
	}
	return &AlternativeTransferCharacteristics{PreferredTransferCharacteristics: m.Payload[0]}, nil
}
