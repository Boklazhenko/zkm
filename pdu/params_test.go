package pdu

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewMandatoryParams(t *testing.T) {
	tests := []struct {
		id    Id
		names []Name
	}{
		{BindTransmitter, []Name{SystemID, Password, SystemType, InterfaceVersion, AddrTON, AddrNPI, AddressRange}},
		{BindReceiver, []Name{SystemID, Password, SystemType, InterfaceVersion, AddrTON, AddrNPI, AddressRange}},
		{BindTransceiver, []Name{SystemID, Password, SystemType, InterfaceVersion, AddrTON, AddrNPI, AddressRange}},
		{BindTransmitterResp, []Name{SystemID}},
		{BindReceiverResp, []Name{SystemID}},
		{BindTransceiverResp, []Name{SystemID}},
		{Outbind, []Name{SystemID, Password}},
		{Unbind, []Name{}},
		{UnbindResp, []Name{}},
		{GenericNack, []Name{}},
		{SubmitSm, []Name{ServiceType, SourceAddrTON, SourceAddrNPI, SourceAddr, DestAddrTON, DestAddrNPI, DestinationAddr,
			ESMClass, ProtocolID, PriorityFlag, ScheduleDeliveryTime, ValidityPeriod, RegisteredDelivery, ReplaceIfPresentFlag,
			DataCoding, SMDefaultMsgID, SMLength, ShortMessage}},
		{SubmitSmResp, []Name{MessageID}},
		{DeliverSm, []Name{ServiceType, SourceAddrTON, SourceAddrNPI, SourceAddr, DestAddrTON, DestAddrNPI, DestinationAddr,
			ESMClass, ProtocolID, PriorityFlag, ScheduleDeliveryTime, ValidityPeriod, RegisteredDelivery, ReplaceIfPresentFlag,
			DataCoding, SMDefaultMsgID, SMLength, ShortMessage}},
		{DeliverSmResp, []Name{MessageID}},
		{DataSm, []Name{ServiceType, SourceAddrTON, SourceAddrNPI, SourceAddr,
			DestAddrTON, DestAddrNPI, DestinationAddr, ESMClass, RegisteredDelivery, DataCoding}},
		{DataSmResp, []Name{MessageID}},
		{QuerySm, []Name{MessageID, SourceAddrTON, SourceAddrNPI, SourceAddr}},
		{QuerySmResp, []Name{MessageID, FinalDate, MessageState, ErrorCode}},
		{CancelSm, []Name{ServiceType, MessageID, SourceAddrTON, SourceAddrNPI, SourceAddr, DestAddrTON, DestAddrNPI, DestinationAddr}},
		{CancelSmResp, []Name{}},
		{ReplaceSm, []Name{MessageID, SourceAddrTON, SourceAddrNPI, SourceAddr, ScheduleDeliveryTime, ValidityPeriod,
			RegisteredDelivery, SMDefaultMsgID, SMLength, ShortMessage}},
		{ReplaceSmResp, []Name{}},
		{EnquireLink, []Name{}},
		{EnquireLinkResp, []Name{}},
		{AlertNotification, []Name{SourceAddrTON, SourceAddrNPI, SourceAddr, EsmeAddrTON, EsmeAddrNPI, EsmeAddr}},
	}

	for _, test := range tests {
		params := NewMandatoryParams(test.id)

		if !reflect.DeepEqual(test.names, params.names) {
			t.Errorf("mandatory param names not equal expected for command id [%v]", test.id)
		}

		if len(params.params) != len(test.names) {
			t.Errorf("length of params not equal len of names for command id [%v]", test.id)
		}

		for _, name := range test.names {
			if _, err := params.Get(name); err != nil {
				t.Errorf("mandatory params for command id [%v] getting error [%v]", test.id, err)
			}
		}
	}
}

func TestOctetStringValue(t *testing.T) {
	lensForNew := []int{0, 4, 10, 60000}

	for _, l := range lensForNew {
		v := NewOctetStringValue(l)
		if v.Len() != l {
			t.Errorf("Len [%v] not equal to the transmitted [%v] after New", v.Len(), l)
		}

		if !reflect.DeepEqual(v.Raw(), make([]byte, l)) {
			t.Errorf("raw data not equal default value for the transmitted length [%v] after New", l)
		}

		if _, err := v.Uint32(); err == nil {
			t.Errorf("Uint32 not return err after New")
		}

		lensForSet := []int{0, 2, 3, 15, 10000}
		for _, l2 := range lensForSet {
			if err := v.Set(make([]byte, l2)); err != nil {
				t.Errorf("set error [%v]", err)
			}

			if v.Len() != l2 {
				t.Errorf("Len %v not equal to the transmitted [%v] after New", v.Len(), l2)
			}

			if !reflect.DeepEqual(v.Raw(), make([]byte, l2)) {
				t.Errorf("raw data not equal default value for the transmitted length [%v] after New", l2)
			}

			if _, err := v.Uint32(); err == nil {
				t.Errorf("Uint32 not return err after New")
			}
		}
	}

	tests := []struct {
		value interface{}
		ok    bool
	}{
		{[]byte{1, 2, 3, 4, 5}, true},
		{"some string", false},
		{4, false},
		{uint8(10), false},
		{uint32(10000), false},
	}

	v := NewOctetStringValue(0)
	for _, test := range tests {
		if err := v.Set(test.value); (err == nil) != test.ok {
			t.Errorf("set return err [%v] for value type [%T]", err, test.value)
		}
	}
}

func TestValue(t *testing.T) {
	type uint32Result struct {
		value uint32
		err   bool
	}
	type expected struct {
		len    int
		raw    []byte
		uint32 uint32Result
	}
	type setting struct {
		value    interface{}
		ok       bool
		expected expected
	}
	type deserialization struct {
		buff        *bytes.Buffer
		ok          bool
		lastBuffLen int
		expected    expected
	}

	tests := []struct {
		createValue     func() Value
		expected        expected
		setting         []setting
		deserialization []deserialization
	}{
		{
			createValue: func() Value {
				return NewOctetStringValue(0)
			},
			expected: expected{
				len: 0,
				raw: []byte{},
				uint32: uint32Result{
					err: true,
				},
			},
			setting: []setting{
				{
					value: []byte{},
					ok:    true,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: []byte{1, 2, 3, 4},
					ok:    true,
					expected: expected{
						len: 4,
						raw: []byte{1, 2, 3, 4},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: "Hello world!",
					ok:    false,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: 155,
					ok:    false,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
			deserialization: []deserialization{
				{
					buff:        bytes.NewBuffer(nil),
					ok:          true,
					lastBuffLen: 0,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					buff:        bytes.NewBuffer([]byte{1, 2, 3, 4}),
					ok:          true,
					lastBuffLen: 4,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
		},
		{
			createValue: func() Value {
				return NewOctetStringValue(4)
			},
			expected: expected{
				len: 4,
				raw: []byte{0, 0, 0, 0},
				uint32: uint32Result{
					err: true,
				},
			},
			setting: []setting{
				{
					value: []byte{},
					ok:    true,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: []byte{1, 2, 3, 4},
					ok:    true,
					expected: expected{
						len: 4,
						raw: []byte{1, 2, 3, 4},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: "Hello world!",
					ok:    false,
					expected: expected{
						len: 4,
						raw: []byte{0, 0, 0, 0},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: 10101,
					ok:    false,
					expected: expected{
						len: 4,
						raw: []byte{0, 0, 0, 0},
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
			deserialization: []deserialization{
				{
					buff:        bytes.NewBuffer([]byte{9, 8, 7, 6}),
					ok:          true,
					lastBuffLen: 0,
					expected: expected{
						len: 4,
						raw: []byte{9, 8, 7, 6},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					buff:        bytes.NewBuffer([]byte{1, 2, 3}),
					ok:          false,
					lastBuffLen: 0,
					expected: expected{
						len: 4,
						raw: []byte{0, 0, 0, 0},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					buff:        bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
					ok:          true,
					lastBuffLen: 1,
					expected: expected{
						len: 4,
						raw: []byte{1, 2, 3, 4},
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
		},
		{
			createValue: func() Value {
				return NewOctetStringValue(200)
			},
			expected: expected{
				len: 200,
				raw: make([]byte, 200),
				uint32: uint32Result{
					err: true,
				},
			},
			setting: []setting{
				{
					value: []byte{},
					ok:    true,
					expected: expected{
						len: 0,
						raw: []byte{},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: []byte{1, 2, 3, 4},
					ok:    true,
					expected: expected{
						len: 4,
						raw: []byte{1, 2, 3, 4},
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: make([]byte, 240),
					ok:    true,
					expected: expected{
						len: 240,
						raw: make([]byte, 240),
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: "data",
					ok:    false,
					expected: expected{
						len: 200,
						raw: make([]byte, 200),
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					value: 0,
					ok:    false,
					expected: expected{
						len: 200,
						raw: make([]byte, 200),
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
			deserialization: []deserialization{
				{
					buff:        bytes.NewBuffer(nil),
					ok:          false,
					lastBuffLen: 0,
					expected: expected{
						len: 200,
						raw: make([]byte, 200),
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					buff:        bytes.NewBuffer(make([]byte, 5)),
					ok:          false,
					lastBuffLen: 0,
					expected: expected{
						len: 200,
						raw: make([]byte, 200),
						uint32: uint32Result{
							err: true,
						},
					},
				},
				{
					buff:        bytes.NewBuffer(make([]byte, 230)),
					ok:          true,
					lastBuffLen: 30,
					expected: expected{
						len: 200,
						raw: make([]byte, 200),
						uint32: uint32Result{
							err: true,
						},
					},
				},
			},
		},
	}

	for i, test := range tests {
		value := test.createValue()
		expected := test.expected
		if value.Len() != expected.len {
			t.Errorf("[%v]: len [%v] of value type [%T] not equals expected [%v] after New", i, value.Len(),
				value, expected.len)
		}

		if !reflect.DeepEqual(value.Raw(), expected.raw) {
			t.Errorf("[%v]: raw [%v] of value type [%T] not equals expected [%v] after New", i, value.Raw(),
				value, expected.raw)
		}

		if v, err := value.Uint32(); (err != nil) != expected.uint32.err {
			t.Errorf("[%v]: uint32 err [%v] of value type [%T] not equals expected [%v] after New", i, err,
				value, expected.uint32.err)
		} else if !expected.uint32.err && v != expected.uint32.value {
			t.Errorf("[%v]: uint32 value [%v] of value type [%T] not equals expected [%v] after New", i, v, value,
				expected.uint32.value)
		}

		for j, setting := range test.setting {
			value := test.createValue()
			expected := setting.expected
			if err := value.Set(setting.value); (err == nil) != setting.ok {
				t.Errorf("[%v.%v]: setting error [%v] of value type [%T] not equals expected [%v]", i, j, err,
					value, setting.ok)
			}

			if value.Len() != expected.len {
				t.Errorf("[%v.%v]: len [%v] of value type [%T] not equals expected [%v] after Set", i, j, value.Len(),
					value, expected.len)
			}

			if !reflect.DeepEqual(value.Raw(), expected.raw) {
				t.Errorf("[%v.%v]: raw [%v] of value type [%T] not equals expected [%v] after Set", i, j, value.Raw(),
					value, expected.raw)
			}

			if v, err := value.Uint32(); (err != nil) != expected.uint32.err {
				t.Errorf("[%v.%v]: uint32 err [%v] of value type [%T] not equals expected [%v] after Set", i, j, err,
					value, expected.uint32.err)
			} else if !expected.uint32.err && v != expected.uint32.value {
				t.Errorf("[%v.%v]: uint32 value [%v] of value type [%T] not equals expected [%v] after Set", i, j,
					v, value, expected.uint32.value)
			}
		}

		for j, deserialization := range test.deserialization {
			value := test.createValue()
			expected := deserialization.expected
			if err := value.Deserialize(deserialization.buff); (err == nil) != deserialization.ok {
				t.Errorf("[%v.%v]: deserialize error [%v] of value type [%T] not equals expected [%v]", i, j, err,
					value, deserialization.ok)
			}

			if deserialization.buff.Len() != deserialization.lastBuffLen {
				t.Errorf("[%v.%v]: last buff len [%v] not equals expected [%v] after Deserialize", i, j,
					deserialization.buff.Len(), deserialization.lastBuffLen)
			}

			if value.Len() != expected.len {
				t.Errorf("[%v.%v]: len [%v] of value type [%T] not equals expected [%v] after Deserialize", i, j,
					value.Len(), value, expected.len)
			}

			if !reflect.DeepEqual(value.Raw(), expected.raw) {
				t.Errorf("[%v.%v]: raw [%v] of value type [%T] not equals expected [%v] after Deserialize", i, j, value.Raw(),
					value, expected.raw)
			}

			if v, err := value.Uint32(); (err != nil) != expected.uint32.err {
				t.Errorf("[%v.%v]: uint32 err [%v] of value type [%T] not equals expected [%v] after Deseriazlie", i, j, err,
					value, expected.uint32.err)
			} else if !expected.uint32.err && v != expected.uint32.value {
				t.Errorf("[%v.%v]: uint32 value [%v] of value type [%T] not equals expected [%v] after Deserialize", i, j, v, value,
					expected.uint32.value)
			}
		}
	}
}
