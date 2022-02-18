package sms

type UserDataHeader struct {
	TotalNumber int
	Sequence    int
	Tag         int
}

func (udh *UserDataHeader) ReadFrom(octets []byte) error {
	octetsLng := len(octets)
	headerLng := int(octets[0]) + 1
	if (octetsLng-headerLng) <= 0 || headerLng <= 5 {
		return ErrIncorrectUserDataHeaderLength
	}

	h := octets[:headerLng]
	udh.Sequence = int(h[5])
	udh.TotalNumber = int(h[4])
	udh.Tag = int(h[3])

	return nil
}
