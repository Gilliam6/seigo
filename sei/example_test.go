package sei_test

import (
	"encoding/hex"
	"fmt"

	"github.com/Gilliam6/seigo/sei"
)

// ExampleNALU_EncodeNALU demonstrates building a SEI NAL unit that carries a
// recovery_point and a content_light_level_info message and then parsing it
// back.
func ExampleNALU_EncodeNALU() {
	rp, _ := (&sei.RecoveryPoint{
		RecoveryFrameCnt: 0,
		ExactMatchFlag:   true,
	}).Encode()
	cll, _ := (&sei.ContentLightLevelInfo{
		MaxContentLightLevel:    4000,
		MaxPicAverageLightLevel: 1000,
	}).Encode()

	out := &sei.NALU{Messages: []sei.Message{rp, cll}}
	naluBytes, err := out.EncodeNALU()
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(naluBytes))

	parsed, err := sei.ParseNALU(naluBytes)
	if err != nil {
		panic(err)
	}
	for _, m := range parsed.Messages {
		fmt.Printf("- %s (%d bytes)\n", m.Type, len(m.Payload))
	}
	// Output:
	// 060601c490040fa003e880
	// - recovery_point (1 bytes)
	// - content_light_level_info (4 bytes)
}

// ExampleParseNALU shows how to dispatch on the payload type after parsing.
func ExampleParseNALU() {
	// A SEI NAL unit carrying a single user_data_unregistered message.
	uuid := [16]byte{0xdc, 0x45, 0xe9, 0xbd, 0xe6, 0xd9, 0x48, 0xb7,
		0x96, 0x2c, 0xd8, 0x20, 0xd9, 0x23, 0xee, 0xef}
	udu := sei.UserDataUnregistered{UUID: uuid, Data: []byte("hello")}
	msg, _ := udu.Encode()
	enc, _ := (&sei.NALU{Messages: []sei.Message{msg}}).EncodeNALU()

	nalu, err := sei.ParseNALU(enc)
	if err != nil {
		panic(err)
	}
	for _, m := range nalu.Messages {
		switch m.Type {
		case sei.PayloadTypeUserDataUnregistered:
			parsed, _ := sei.ParseUserDataUnregistered(m)
			fmt.Printf("UUID=%x payload=%q\n", parsed.UUID, parsed.Data)
		default:
			fmt.Printf("unhandled: %s\n", m.Type)
		}
	}
	// Output:
	// UUID=dc45e9bde6d948b7962cd820d923eeef payload="hello"
}
