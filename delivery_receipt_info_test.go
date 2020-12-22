package zkm

import (
	"encoding/hex"
	"testing"
)

func TestNewDeliveryReceiptInfo(t *testing.T) {
	if b, err := hex.DecodeString("000000D70000000500000000000000030001013739353030383932353638000001373737000400000000000" +
		"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
		"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
		"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
		"D376464323464303361373136000427000102"); err == nil {

		pdu := NewEmptyPdu()
		if err = pdu.Deserialize(b); err == nil {
			dri := NewDeliveryReceiptInfo(pdu)

			if dri.Id != "cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716" {
				t.Errorf("id [%v] not equals expected [%v]", dri.Id, "cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716")
			}

			if dri.State != 2 {
				t.Errorf("state [%v] not equals expected [%v]", dri.State, 2)
			}

			if dri.Err != 0 {
				t.Errorf("err [%v] not equals expected [%v]", dri.Err, 0)
			}

			expectedText := "id:cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716 sub:001 dlvrd:001 submit date:2010241205 done date:2010241206 stat:DELIVRD err:000"
			if dri.Text != expectedText {
				t.Errorf("text [%v] not equals expected [%v]", dri.Text, expectedText)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}
