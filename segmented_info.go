package zkm

type SegmentedInfo struct {
	part       int
	totalParts int
	id         int
}

func NewSegmentedInfo(pdu *Pdu) *SegmentedInfo {
	segmentedInfo := &SegmentedInfo{part: 1, totalParts: 1, id: 0}

	esmClass, err := pdu.GetMainAsUint32(ESMClass)

	if err != nil {
		return segmentedInfo
	}

	if (esmClass & 0x40) != 0 {
		sm, err := pdu.GetMainAsRaw(ShortMessage)

		smLength := len(sm)

		if err != nil || smLength == 0 {
			return segmentedInfo
		}

		udhl := int(sm[0]) + 1
		if smLength < udhl {
			udhl = smLength
		}

		for i := 1; i < udhl; {
			iei := int(sm[i])
			i++
			iel := int(sm[i])
			i++

			switch iei {
			case 0x00: //concat 8bit
				if iel == 3 {
					segmentedInfo.id = int(sm[i])
					i++
					segmentedInfo.totalParts = int(sm[i])
					i++
					segmentedInfo.part = int(sm[i])
					i++
					return segmentedInfo
				}
			case 0x08: //concat 16bit
				if iel == 4 {
					segmentedInfo.id = (int(sm[i]) << 8) | (int(sm[i+1]))
					i += 2
					segmentedInfo.totalParts = int(sm[i])
					i++
					segmentedInfo.part = int(sm[i])
					i++
					return segmentedInfo
				}
			default:
				i += iel
			}
		}
	} else if sarTotalSegments, err := pdu.GetOptAsUint32(TagSarTotalSegments); err == nil {
		sarMsgRefNum, err := pdu.GetOptAsUint32(TagSarMsgRefNum)

		if err != nil {
			return segmentedInfo
		}

		sarSegmentSeqnum, err := pdu.GetOptAsUint32(TagSarSegmentSeqnum)

		if err != nil {
			return segmentedInfo
		}

		segmentedInfo.id = int(sarMsgRefNum)
		segmentedInfo.totalParts = int(sarTotalSegments)
		segmentedInfo.part = int(sarSegmentSeqnum)
	}

	return segmentedInfo
}
