package zkm

func CreateSubmits(text string, msgRefNumProvider func() uint16) ([]*Pdu, error) {
	return createPdus(text, msgRefNumProvider, SubmitSm)
}

func CreateDeliveries(text string, msgRefNumProvider func() uint16) ([]*Pdu, error) {
	return createPdus(text, msgRefNumProvider, DeliverSm)
}

func createPdus(text string, msgRefNumProvider func() uint16, id Id) ([]*Pdu, error) {
	pdus := make([]*Pdu, 0)
	dcs := SmscDefaultAlphabetScheme
	maxLen := 160
	cutLen := 153
	b := make([]byte, 0)
	var err error

	if b, err = Encode(text, Gsm7Unpacked()); err != nil {
		if b, err = Encode(text, Ucs2()); err == nil {
			dcs = Ucs2Scheme
			maxLen = 140
			cutLen = 132
		} else {
			return nil, err
		}
	}

	if len(b) <= maxLen {
		pdu := NewPdu(id)
		if err = pdu.SetMain(DataCoding, dcs); err != nil {
			return nil, err
		}
		if err = pdu.SetMain(ShortMessage, b); err != nil {
			return nil, err
		}
		if err = pdu.SetMain(SMLength, len(b)); err != nil {
			return nil, err
		}
		pdus = append(pdus, pdu)
	} else {
		countParts := (len(b)-1)/cutLen + 1
		msgRefNum := msgRefNumProvider()
		for i := 0; i < countParts; i++ {
			pdu := NewPdu(id)
			udh := make([]byte, 7)
			udh[0] = 0x06                  // length of user data header
			udh[1] = 0x08                  // information element identifier, CSMS 16 bit reference number
			udh[2] = 0x04                  // length of remaining header
			udh[3] = uint8(msgRefNum >> 8) // most significant byte of the reference number
			udh[4] = uint8(msgRefNum)      // least significant byte of the reference number
			udh[5] = uint8(countParts)     // total number of message parts
			udh[6] = uint8(i + 1)          // current message part

			sm := make([]byte, 0)
			if i != countParts-1 {
				sm = append(udh, b[i*cutLen:(i+1)*cutLen]...)
			} else {
				sm = append(udh, b[i*cutLen:]...)
			}

			if err = pdu.SetMain(DataCoding, dcs); err != nil {
				return nil, err
			}
			if err = pdu.SetMain(ESMClass, 0x40); err != nil {
				return nil, err
			}
			if err = pdu.SetMain(ShortMessage, sm); err != nil {
				return nil, err
			}
			if err = pdu.SetMain(SMLength, len(sm)); err != nil {
				return nil, err
			}
			pdus = append(pdus, pdu)
		}
	}

	return pdus, nil
}
