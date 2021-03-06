package zkm

import (
	"encoding/hex"
	"testing"
)

func TestNewSegmentedInfo(t *testing.T) {
	tests := []struct {
		rawPdu             string
		expectedId         int
		expectedPart       int
		expectedTotalParts int
	}{
		{
			rawPdu:             "0000004A000000040000000000000003000001373737000101373939393833313533383500000000003230313130323230303832373030302B00010000000B48656C6C6F20576F726C64",
			expectedId:         0,
			expectedPart:       1,
			expectedTotalParts: 1,
		},
		{
			rawPdu:             "000000CB000000040000000000000003000001373737000101373932393036373139353200400000003230313130323230313135323030302B00010000008C0608044CAC0201432B2B20D188D0B8D180D0BED0BAD0BE20D0B8D181D0BFD0BED0BBD18CD0B7D183D0B5D182D181D18F20D0B4D0BBD18F20D180D0B0D0B7D180D0B0D0B1D0BED182D0BAD0B820D0BFD180D0BED0B3D180D0B0D0BCD0BCD0BDD0BED0B3D0BE20D0BED0B1D0B5D181D0BFD0B5D187D0B5D0BDD0B8D18F2C20D18FD0B2D0BBD18FD18FD181D18C",
			expectedId:         19628,
			expectedPart:       1,
			expectedTotalParts: 2,
		},
		{
			rawPdu:             "000000A4000000040000000000000004000001373737000101373932393036373139353200400000003230313130323230313135323030302B0001000000650608044CAC020220D0BED0B4D0BDD0B8D0BC20D0B8D0B720D181D0B0D0BCD18BD18520D0BFD0BED0BFD183D0BBD18FD180D0BDD18BD18520D18FD0B7D18BD0BAD0BED0B220D0BFD180D0BED0B3D180D0B0D0BCD0BCD0B8D180D0BED0B2D0B0D0BDD0B8D18F",
			expectedId:         19628,
			expectedPart:       2,
			expectedTotalParts: 2,
		},
		{
			rawPdu:             "000000D4000000040000000000000003000001373737000101373939393634313533313200000000003230313130323230313533303030302B000100000085432B2B20D188D0B8D180D0BED0BAD0BE20D0B8D181D0BFD0BED0BBD18CD0B7D183D0B5D182D181D18F20D0B4D0BBD18F20D180D0B0D0B7D180D0B0D0B1D0BED182D0BAD0B820D0BFD180D0BED0B3D180D0B0D0BCD0BCD0BDD0BED0B3D0BE20D0BED0B1D0B5D181D0BFD0B5D187D0B5D0BDD0B8D18F2C20D18FD0B2D0BBD18FD18FD181D18C020C00026343020E000102020F000101",
			expectedId:         25411,
			expectedPart:       1,
			expectedTotalParts: 2,
		},
		{
			rawPdu:             "000000AD000000040000000000000004000001373737000101373939393634313533313200000000003230313130323230313533303030302B00010000005E20D0BED0B4D0BDD0B8D0BC20D0B8D0B720D181D0B0D0BCD18BD18520D0BFD0BED0BFD183D0BBD18FD180D0BDD18BD18520D18FD0B7D18BD0BAD0BED0B220D0BFD180D0BED0B3D180D0B0D0BCD0BCD0B8D180D0BED0B2D0B0D0BDD0B8D18F020C00026343020E000102020F000102",
			expectedId:         25411,
			expectedPart:       2,
			expectedTotalParts: 2,
		},
	}

	for _, test := range tests {
		b, err := hex.DecodeString(test.rawPdu)

		if err != nil {
			panic(err)
		}

		pdu := NewEmptyPdu()
		err = pdu.Deserialize(b)

		if err != nil {
			panic(err)
		}

		segmentedInfo := NewSegmentedInfo(pdu)

		if segmentedInfo.Id != test.expectedId {
			t.Errorf("in segmented info for pdu %v id %v not equals expected %v",
				test.rawPdu, segmentedInfo.Id, test.expectedId)
		}

		if segmentedInfo.Part != test.expectedPart {
			t.Errorf("in segmented info for pdu %v part %v not equals expected %v",
				test.rawPdu, segmentedInfo.Part, test.expectedPart)
		}

		if segmentedInfo.TotalParts != test.expectedTotalParts {
			t.Errorf("in segmented info for pdu %v total parts %v not equals expected %v",
				test.rawPdu, segmentedInfo.TotalParts, test.expectedTotalParts)
		}
	}
}
