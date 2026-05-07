package bits

// AddEmulationPrevention encodes an RBSP byte sequence into an EBSP byte
// sequence by inserting an emulation_prevention_three_byte (0x03) wherever a
// pattern of two consecutive 0x00 bytes is followed by a byte with a value in
// {0x00, 0x01, 0x02, 0x03}, per §7.4.1.1 of the H.264 specification.
//
// This is the exact inverse of RemoveEmulationPrevention. A well-formed SEI
// RBSP terminates with rbsp_trailing_bits equal to 0x80, so the resulting
// EBSP never ends with a 0x00 byte; constraints of that sort are the
// responsibility of whoever produced the RBSP.
func AddEmulationPrevention(rbsp []byte) []byte {
	out := make([]byte, 0, len(rbsp)+len(rbsp)/256)
	zeros := 0
	for _, b := range rbsp {
		if zeros >= 2 && b <= 0x03 {
			out = append(out, 0x03)
			zeros = 0
		}
		out = append(out, b)
		if b == 0 {
			zeros++
		} else {
			zeros = 0
		}
	}
	return out
}

// RemoveEmulationPrevention decodes an EBSP byte sequence back into an RBSP
// byte sequence by stripping every emulation_prevention_three_byte (0x03)
// that follows two or more 0x00 bytes.
//
// Bytes that do not match the prevention pattern are passed through verbatim,
// matching the decoder behavior described in §7.4.1.1.
func RemoveEmulationPrevention(ebsp []byte) []byte {
	out := make([]byte, 0, len(ebsp))
	zeros := 0
	for i := 0; i < len(ebsp); i++ {
		b := ebsp[i]
		if zeros >= 2 && b == 0x03 {
			zeros = 0
			continue
		}
		out = append(out, b)
		if b == 0 {
			zeros++
		} else {
			zeros = 0
		}
	}
	return out
}
