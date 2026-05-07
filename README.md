# seigo

Parser/generator of H.264 SEI (Supplemental Enhancement Information) NAL units in pure Go.

The implementation follows ITU-T Rec. H.264 / ISO-IEC 14496-10, in particular:

- §7.3.2.3 (`sei_rbsp`) and §7.3.2.3.1 (`sei_message`) for the message envelope.
- §7.4.1.1 for emulation-prevention byte handling between RBSP and EBSP.
- §9.1 / §9.1.1 for unsigned and signed Exp-Golomb codes.
- Annex D for individual SEI message payloads.

## Install

```bash
go get github.com/Gilliam6/seigo
```

Requires Go 1.21 or newer.

## Packages

- `github.com/Gilliam6/seigo/bits` — bit reader/writer, Exp-Golomb codes, RBSP/EBSP conversion. Useful on its own when working with any H.264 syntax structure.
- `github.com/Gilliam6/seigo/sei` — SEI NAL unit envelope, generic message access, and typed encoder/decoders for the most common payload types.

## Generating a SEI NAL unit

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/Gilliam6/seigo/sei"
)

func main() {
	rp, _ := (&sei.RecoveryPoint{
		RecoveryFrameCnt: 0,
		ExactMatchFlag:   true,
	}).Encode()
	cll, _ := (&sei.ContentLightLevelInfo{
		MaxContentLightLevel:    4000,
		MaxPicAverageLightLevel: 1000,
	}).Encode()

	nalu := &sei.NALU{Messages: []sei.Message{rp, cll}}
	bytes, err := nalu.EncodeNALU() // includes 1-byte NAL header (0x06)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bytes))
}
```

`EncodeNALU` returns the full NAL unit, ready to be wrapped in an Annex B
start-code prefix or in a length-prefixed AVCC sample. Use `EncodeEBSP` if you
want only the body (no header), or `EncodeRBSP` to get the unescaped RBSP.

## Parsing a SEI NAL unit

```go
nalu, err := sei.ParseNALU(naluBytes)
if err != nil {
	return err
}
for _, m := range nalu.Messages {
	switch m.Type {
	case sei.PayloadTypeRecoveryPoint:
		rp, _ := sei.ParseRecoveryPoint(m)
		fmt.Println("recovery_frame_cnt =", rp.RecoveryFrameCnt)
	case sei.PayloadTypeUserDataUnregistered:
		udu, _ := sei.ParseUserDataUnregistered(m)
		fmt.Printf("UUID=%x payload=% x\n", udu.UUID, udu.Data)
	default:
		// Raw payload bytes are always available for unsupported types.
		fmt.Printf("%s: %d bytes\n", m.Type, len(m.Payload))
	}
}
```

`ParseEBSP` and `ParseRBSP` are also available when you have already stripped
the NAL unit header or the emulation prevention bytes.

## Typed messages

The following payloads have first-class encode/decode helpers:

| Payload type | Constant | Struct |
| --- | --- | --- |
| 2 | `PayloadTypePanScanRect` | `PanScanRect` |
| 3 | `PayloadTypeFillerPayload` | `FillerPayload` |
| 4 | `PayloadTypeUserDataRegisteredITUTT35` | `UserDataRegisteredITUTT35` |
| 5 | `PayloadTypeUserDataUnregistered` | `UserDataUnregistered` |
| 6 | `PayloadTypeRecoveryPoint` | `RecoveryPoint` |
| 45 | `PayloadTypeFramePackingArrangement` | `FramePackingArrangement` |
| 47 | `PayloadTypeDisplayOrientation` | `DisplayOrientation` |
| 137 | `PayloadTypeMasteringDisplayColourVolume` | `MasteringDisplayColourVolume` |
| 144 | `PayloadTypeContentLightLevelInfo` | `ContentLightLevelInfo` |
| 147 | `PayloadTypeAlternativeTransferCharacteristics` | `AlternativeTransferCharacteristics` |

For other payload types the library still performs framing (it parses the
payload type and size and slices the payload bytes), so you can read or
generate any SEI message at the byte level.

`buffering_period` and `pic_timing` parsing depend on the SPS / VUI HRD
parameters and are therefore deferred to the caller; both can still be carried
through the framework by working directly with `Message` payload bytes.

## Tests

```bash
go test ./...
go test -race -cover ./...
```

## License

MIT — see [LICENSE](LICENSE).
