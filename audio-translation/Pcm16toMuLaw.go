package audio_translator

// μ-law encoding lookup table (G.711 standard), this might be useless, todo
var muLawTable [256]int16

func init() {
	for i := 0; i < 256; i++ {
		muLawTable[i] = int16(((i << 3) - 128) << 2)
	}
}

// ConvertPCM16ToMuLaw converts 16-bit PCM audio to μ-law (G.711), what twilo needs
func ConvertPCM16ToMuLaw(input []byte) []byte {
	output := make([]byte, len(input)/2)
	for i := 0; i < len(input); i += 2 {
		// Read 16-bit PCM sample (little-endian)
		sample := int16(input[i]) | int16(input[i+1])<<8

		// Convert to μ-law
		output[i/2] = LinearToMuLaw(sample)
	}
	return output
}

// LinearToMuLaw converts a 16-bit PCM sample to 8-bit μ-law (G.711)
func LinearToMuLaw(sample int16) byte {
	const (
		bias = 0x84  // Bias for μ-law encoding
		clip = 32635 // Maximum amplitude
	)

	// Get sign and make sample absolute
	sign := byte(0x00)
	if sample < 0 {
		sample = -sample
		sign = 0x80
	}

	// Clip sample to maximum range
	if sample > clip {
		sample = clip
	}

	// Add bias to avoid distortion
	sample += bias

	// Find exponent (log2 of sample)
	exponent := byte(7)
	expMask := int16(0x4000) // 0x4000 = 2^14
	for sample&expMask == 0 && exponent > 0 {
		exponent--
		expMask >>= 1
	}

	// Extract mantissa (next 4 bits)
	mantissa := (sample >> (exponent + 3)) & 0x0F

	// Combine sign, exponent, and mantissa
	ulawByte := ^(sign | (exponent << 4) | byte(mantissa))

	return ulawByte
}
