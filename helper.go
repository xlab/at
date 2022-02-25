package at

import "strconv"

func parseUint8(str string) (uint8, error) {
	i, err := strconv.ParseUint(str, 10, 8)
	return uint8(i), err
}

func parseUint16(str string) (uint16, error) {
	i, err := strconv.ParseUint(str, 10, 16)
	return uint16(i), err
}
