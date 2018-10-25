package pdu

import "fmt"

// Swap semi-octets in octet
func Swap(octet byte) byte {
	return (octet << 4) | (octet >> 4 & 0x0F)
}

// Encode to semi-octets
func Encode(value int) byte {
	lo := byte(value % 10)
	hi := byte((value % 100) / 10)
	return hi<<4 | lo
}

// Decode form semi-octets
func Decode(octet byte) int {
	lo := octet & 0x0F
	hi := octet >> 4 & 0x0F
	return int(hi)*10 + int(lo)
}

// EncodeSemi packs the given numerical chunks in a semi-octet
// representation as described in 3GPP TS 23.040.
func EncodeSemi(chunks ...uint64) []byte {
	digits := make([]uint8, 0, len(chunks))
	for _, c := range chunks {
		var bucket []uint8
		if c < 10 {
			digits = append(digits, 0)
		}
		for c > 0 {
			d := c % 10
			bucket = append(bucket, uint8(d))
			c = (c - d) / 10
		}
		for i := range bucket {
			digits = append(digits, bucket[len(bucket)-1-i])
		}
	}
	octets := make([]byte, 0, len(digits)/2+1)
	for i := 0; i < len(digits); i += 2 {
		if len(digits)-i < 2 {
			octets = append(octets, 0xF0|digits[i])
			return octets
		}
		octets = append(octets, digits[i+1]<<4|digits[i])
	}
	return octets
}

// DecodeSemi unpacks numerical chunks from the given semi-octet encoded data.
func DecodeSemi(octets []byte) []int {
	chunks := make([]int, 0, len(octets)*2)
	for _, oct := range octets {
		half := oct >> 4
		if half == 0xF {
			chunks = append(chunks, int(oct&0x0F))
			return chunks
		}
		chunks = append(chunks, int(oct&0x0F)*10+int(half))
	}
	return chunks
}

// DecodeSemiAddress unpacks phone numbers from the given semi-octet encoded data.
// This method is different from DecodeSemi because a 0x00 byte should be interpreted as
// two distinct digits. There 0x00 will be "00".
func DecodeSemiAddress(octets []byte) (str string) {
	for _, oct := range octets {
		half := oct >> 4
		if half == 0xF {
			str += fmt.Sprintf("%d", oct&0x0F)
			return
		}
		str += fmt.Sprintf("%d%d", oct&0x0F, half)
	}
	return
}
