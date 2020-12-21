package zkm

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

func BenchmarkPduDeserialize(b *testing.B) {
	raw, err := hex.DecodeString("000000D70000000500000000000000030001013739353030383932353638000001373737000400000000000" +
		"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
		"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
		"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
		"D376464323464303361373136000427000102")

	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdu := NewEmptyPdu()
		if err := pdu.Deserialize(raw); err != nil {
			panic(err)
		}
	}
}

func TestPdu(t *testing.T) {
	type header struct {
		l      uint32
		id     Id
		status Status
		seq    uint32
	}
	type mandatoryParams []struct {
		name  Name
		value interface{}
	}
	type optionalParams []struct {
		tag   Tag
		value interface{}
	}

	createPdu := func(header header, mandatoryParams mandatoryParams, optionalParams optionalParams) (*Pdu, error) {
		pdu := NewPdu(header.id)
		pdu.SetStatus(header.status)
		pdu.SetSeq(header.seq)
		for _, p := range mandatoryParams {
			if mp, err := pdu.mandatoryParams.get(p.name); err != nil {
				return nil, err
			} else {
				mp.value().set(p.value)
			}
		}

		for _, p := range optionalParams {
			if err := pdu.optionalParams.add(p.tag, p.value); err != nil {
				return nil, err
			}
		}

		return pdu, nil
	}

	assertHeader := func(pdu *Pdu, expected header) error {
		if pdu.Len() != expected.l {
			return fmt.Errorf("len [%v] not equals expected [%v]", pdu.Len(), expected.l)
		}

		if pdu.Id() != expected.id {
			return fmt.Errorf("id [%v] not equals expected [%v]", pdu.Id(), expected.id)
		}

		if pdu.Status() != expected.status {
			return fmt.Errorf("status [%v] not equals expected [%v]", pdu.Status(), expected.status)
		}

		if pdu.Seq() != expected.seq {
			return fmt.Errorf("seq [%v] not equals expected [%v]", pdu.Seq(), expected.seq)
		}

		return nil
	}
	assertMandatoryParams := func(pdu *Pdu, expected mandatoryParams) error {
		for _, e := range expected {
			if p, err := pdu.mandatoryParams.get(e.name); err != nil {
				return fmt.Errorf("[%v] %v", e.name, err)
			} else {
				equals := false
				switch v := e.value.(type) {
				case []byte:
					equals = reflect.DeepEqual(p.value().raw(), v)
				case string:
					equals = p.value().String() == v
				case int:
					if u, err := p.value().uint32(); err == nil {
						equals = u == uint32(v)
					}
				}

				if !equals {
					return fmt.Errorf("[%v] value %v not equals expected [%v]", e.name, p.value(), e.value)
				}
			}
		}
		return nil
	}
	assertOptionalParams := func(pdu *Pdu, expected optionalParams) error {
		for _, e := range expected {
			if p, err := pdu.optionalParams.get(e.tag); err != nil {
				return fmt.Errorf("[%v] %v", e.tag, err)
			} else {
				equals := false
				switch v := e.value.(type) {
				case []byte:
					equals = reflect.DeepEqual(p.value().raw(), v)
				case string:
					equals = p.value().String() == v
				case int:
					if u, err := p.value().uint32(); err == nil {
						equals = u == uint32(v)
					}
				}

				if !equals {
					return fmt.Errorf("[%v] value %v not equals expected [%v]", e.tag, p.value(), e.value)
				}
			}
		}
		return nil
	}
	assert := func(pdu *Pdu, expectedHeader header, expectedMandatoryParams mandatoryParams,
		expectedOptionalParams optionalParams, b []byte) error {
		if err := assertHeader(pdu, expectedHeader); err != nil {
			return err
		}

		if err := assertMandatoryParams(pdu, expectedMandatoryParams); err != nil {
			return err
		}

		if err := assertOptionalParams(pdu, expectedOptionalParams); err != nil {
			return err
		}

		if !reflect.DeepEqual(pdu.Serialize(), b) {
			return fmt.Errorf("serialize [%v] not equals expected [%v]", pdu.Serialize(), b)
		}

		return nil
	}
	createBytes := func(s string) []byte {
		if b, err := hex.DecodeString(s); err == nil {
			return b
		} else {
			panic(err)
		}
	}

	tests := []struct {
		id              int
		b               []byte
		ok              bool
		header          header
		mandatoryParams mandatoryParams
		optionalParams  optionalParams
	}{
		{
			id: 1,
			b:  createBytes("0000002A0000000200000000000000026175746F5F636C69656E740070617373776F7264000034000000"),
			ok: true,
			header: header{
				l:      42,
				id:     BindTransmitter,
				status: EsmeROk,
				seq:    2,
			},
			mandatoryParams: mandatoryParams{
				{AddrTON, 0},
				{AddrNPI, 0},
				{AddressRange, ""},
				{SystemID, "auto_client"},
				{Password, "password"},
				{SystemType, ""},
				{InterfaceVersion, 52},
			},
		},
		{
			id: 2,
			b:  createBytes("00000016800000010000000000000001000210000134"),
			ok: true,
			header: header{
				l:      22,
				id:     BindReceiverResp,
				status: EsmeROk,
				seq:    1,
			},
			mandatoryParams: mandatoryParams{
				{SystemID, ""},
			},
			optionalParams: optionalParams{
				{TagScInterfaceVersion, 52},
			},
		},
		{
			id: 3,
			b: createBytes("0000004A00000004000000000000000300000137373700010137393530303839323536380000000000323031" +
				"3032343132303630343030302B00010000000B48656C6C6F20576F726C64"),
			ok: true,
			header: header{
				l:      74,
				id:     SubmitSm,
				status: EsmeROk,
				seq:    3,
			},
			mandatoryParams: mandatoryParams{
				{SourceAddrTON, 0},
				{SourceAddr, "777"},
				{ScheduleDeliveryTime, ""},
				{ReplaceIfPresentFlag, 0},
				{SMLength, 11},
				{DestAddrTON, 1},
				{ShortMessage, []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100}},
				{ServiceType, ""},
				{SourceAddrNPI, 1},
				{DestinationAddr, "79500892568"},
				{ESMClass, 0},
				{ProtocolID, 0},
				{SMDefaultMsgID, 0},
				{DestAddrNPI, 1},
				{PriorityFlag, 0},
				{ValidityPeriod, "201024120604000+"},
				{RegisteredDelivery, 1},
				{DataCoding, 0},
			},
		},
		{
			id: 4,
			b: createBytes("0000003580000004000000000000000363623963343066312D306161312D346233652D616662382D37646432" +
				"346430336137313600"),
			ok: true,
			header: header{
				l:      53,
				id:     SubmitSmResp,
				status: EsmeROk,
				seq:    3,
			},
			mandatoryParams: mandatoryParams{
				{MessageID, "cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716"},
			},
		},
		{
			id: 5,
			b: createBytes("000000D70000000500000000000000030001013739353030383932353638000001373737000400000000000" +
				"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
				"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
				"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
				"D376464323464303361373136000427000102"),
			ok: true,
			header: header{
				l:      215,
				id:     DeliverSm,
				status: EsmeROk,
				seq:    3,
			},
			mandatoryParams: mandatoryParams{
				{DestinationAddr, "777"},
				{ReplaceIfPresentFlag, 0},
				{DataCoding, 1},
				{ShortMessage, []byte("id:cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716 sub:001 dlvrd:001 submit " +
					"date:2010241205 done date:2010241206 stat:DELIVRD err:000")},
				{DestAddrTON, 0},
				{DestAddrNPI, 1},
				{ProtocolID, 0},
				{PriorityFlag, 0},
				{ScheduleDeliveryTime, ""},
				{ValidityPeriod, ""},
				{SMLength, 122},
				{ServiceType, ""},
				{SourceAddr, "79500892568"},
				{SMDefaultMsgID, 0},
				{ESMClass, 4},
				{RegisteredDelivery, 0},
				{SourceAddrTON, 1},
				{SourceAddrNPI, 1},
			},
			optionalParams: optionalParams{
				{TagReceiptedMessageID, "cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716"},
				{TagMessageStateOption, 2},
			},
		},
		{
			id: 6,
			b:  createBytes("0000001180000005000000000000000300"),
			ok: true,
			header: header{
				l:      17,
				id:     DeliverSmResp,
				status: EsmeROk,
				seq:    3,
			},
			mandatoryParams: mandatoryParams{
				{MessageID, ""},
			},
		},
		{
			id: 7,
			b:  createBytes("00000010000000150000000000000008"),
			ok: true,
			header: header{
				l:      16,
				id:     EnquireLink,
				status: EsmeROk,
				seq:    8,
			},
		},
		{
			id: 8,
			b:  createBytes("00000010800000150000000000000008"),
			ok: true,
			header: header{
				l:      16,
				id:     EnquireLinkResp,
				status: EsmeROk,
				seq:    8,
			},
		},
		{
			id: 9,
			b:  createBytes("0000001000000006000000000000000A"),
			ok: true,
			header: header{
				l:      16,
				id:     Unbind,
				status: EsmeROk,
				seq:    10,
			},
		},
		{
			id: 10,
			b:  createBytes("0000001080000006000000000000000A"),
			ok: true,
			header: header{
				l:      16,
				id:     UnbindResp,
				status: EsmeROk,
				seq:    10,
			},
		},
		{
			id: 11,
			b: createBytes("000000AF0000000500000000000000030001013739353030383932353638000001373737000400000000000" +
				"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
				"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
				"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
				"D376464323464303361373136000427000102"),
			ok: false,
			header: header{
				l:      215,
				id:     DeliverSm,
				status: EsmeROk,
				seq:    3,
			},
			mandatoryParams: mandatoryParams{
				{DestinationAddr, "777"},
				{ReplaceIfPresentFlag, 0},
				{DataCoding, 1},
				{ShortMessage, []byte("id:cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716 sub:001 dlvrd:001 submit " +
					"date:2010241205 done date:2010241206 stat:DELIVRD err:000")},
				{DestAddrTON, 0},
				{DestAddrNPI, 1},
				{ProtocolID, 0},
				{PriorityFlag, 0},
				{ScheduleDeliveryTime, ""},
				{ValidityPeriod, ""},
				{SMLength, 122},
				{ServiceType, ""},
				{SourceAddr, "79500892568"},
				{SMDefaultMsgID, 0},
				{ESMClass, 4},
				{RegisteredDelivery, 0},
				{SourceAddrTON, 1},
				{SourceAddrNPI, 1},
			},
			optionalParams: optionalParams{
				{TagReceiptedMessageID, "cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716"},
				{TagMessageStateOption, 2},
			},
		},
		{
			id: 12,
			b: createBytes("0000FFF0000000500000000000000030001013739353030383932353638000001373737000400000000000" +
				"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
				"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
				"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
				"D3764643234643"),
			ok: false,
		},
		{
			id: 13,
			b: createBytes("000000AF0000000500000000000000030001013739353030383932353638000001373737000400000000000" +
				"001007A69643A63623963343066312D306161312D346233652D616662382D376464323464303361373136207375623A3030312" +
				"0646C7672643A303031207375626D697420646174653A3230313032343132303520646F6E6520646174653A323031303234313" +
				"2303620737461743A44454C49565244206572723A303030001E002563623963343066312D306161312D346233652D616662382" +
				"D376464323464303361373136000427000102"),
			ok: false,
		},
		{
			id: 14,
			b:  createBytes("00000016800000010000000000000001000210000234"),
			ok: false,
		},
		{
			id: 14,
			b:  createBytes("0000001680000001000000000000000100021000013456"),
			ok: false,
		},
		{
			id: 15,
			b:  createBytes("000000158000000100000000000000010210000134"),
			ok: false,
		},
	}

	for _, test := range tests {
		pdu := NewEmptyPdu()
		if err := pdu.Deserialize(test.b); (err == nil) != test.ok {
			t.Errorf("[%v] err [%v] not equals expected [%v] after deserialize", test.id, err, test.ok)
		}

		if !test.ok {
			continue
		}

		if err := assert(pdu, test.header, test.mandatoryParams, test.optionalParams, test.b); err != nil {
			t.Errorf("[%v] %v after deserialize", test.id, err)
		}

		pdu, err := createPdu(test.header, test.mandatoryParams, test.optionalParams)

		if err != nil {
			t.Errorf("[%v] %v", test.id, err)
		}

		if err := assert(pdu, test.header, test.mandatoryParams, test.optionalParams, test.b); err != nil {
			t.Errorf("[%v] %v after create and set", test.id, err)
		}
	}
}
