package audio_translator

func MuLawToLinear(ulaw byte) int16 {
	const bias = 132 // μ-law bias
	ulaw = ^ulaw     // Invert all bits

	sign := int16(ulaw & 0x80)            // Get sign bit
	exponent := int16((ulaw >> 4) & 0x07) // Extract exponent
	mantissa := int16(ulaw & 0x0F)        // Extract mantissa

	// Proper mantissa scaling
	sample := ((mantissa | 0x10) << 1) + 1

	// Apply exponent shift
	sample <<= (exponent + 2) // Exponent correction

	// Apply sign
	if sign != 0 {
		return -sample
	}
	return sample
}

// ConvertMuLawToPCM16 converts μ-law (G.711) to 16-bit PCM audio, what gcp needs
func ConvertMuLawToPCM16(audioData []byte) []byte {
	linearPCM := make([]byte, len(audioData)*2)
	for i, ulawByte := range audioData {
		sample := MuLawToLinear(ulawByte)
		linearPCM[i*2] = byte(sample & 0xFF)
		linearPCM[i*2+1] = byte((sample >> 8) & 0xFF)
	}
	return linearPCM
}
