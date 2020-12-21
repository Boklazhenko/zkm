package zkm

import (
	"testing"
)

func TestCreatePdus(t *testing.T) {
	type expected struct {
		totalParts int
		esmClass   uint32
		dcs        uint32
	}

	type test struct {
		text              string
		msgRefNumProvider func() uint16
		expected          *expected
	}
	tests := []*test{
		{
			text: "",
			msgRefNumProvider: func() uint16 {
				return 1
			},
			expected: &expected{
				totalParts: 1,
				esmClass:   0,
				dcs:        SmscDefaultAlphabetScheme,
			},
		},
		{
			text: "Hello World",
			msgRefNumProvider: func() uint16 {
				return 3000
			},
			expected: &expected{
				totalParts: 1,
				esmClass:   0,
				dcs:        SmscDefaultAlphabetScheme,
			},
		},
		{
			text: "Привет, Мир",
			msgRefNumProvider: func() uint16 {
				return 200
			},
			expected: &expected{
				totalParts: 1,
				esmClass:   0,
				dcs:        Ucs2Scheme,
			},
		},
		{
			text: "C++ is a high-level, general-purpose programming language created by Bjarne Stroustrup as an extension of the C programming language, or \"C with Classes\". The language has expanded significantly over time, and modern C++ has object-oriented, generic, and functional features in addition to facilities for low-level memory manipulation. It is almost always implemented as a compiled language, and many vendors provide C++ compilers, including the Free Software Foundation, LLVM, Microsoft, Intel, Oracle, and IBM, so it is available on many platforms.",
			msgRefNumProvider: func() uint16 {
				return 255
			},
			expected: &expected{
				totalParts: 4,
				esmClass:   0x40,
				dcs:        SmscDefaultAlphabetScheme,
			},
		},
		{
			text: "C++ широко используется для разработки программного обеспечения, являясь одним из самых популярных языков программирования",
			msgRefNumProvider: func() uint16 {
				return 60000
			},
			expected: &expected{
				totalParts: 2,
				esmClass:   0x40,
				dcs:        Ucs2Scheme,
			},
		},
	}

	for _, id := range []Id{SubmitSm, DeliverSm} {
		for _, test := range tests {
			pdus, err := createPdus(test.text, test.msgRefNumProvider, id)

			if err != nil {
				t.Fatalf("failed create pdus with id %v: %v", id, err)
			}

			if len(pdus) != test.expected.totalParts {
				t.Fatalf("with text, %v len of pdus %v not equals expected %v", test.text, len(pdus), test.expected.totalParts)
			}

			for i, pdu := range pdus {
				if pdu.Id() != id {
					t.Fatalf("with text %v, id %v not equals expected %v", test.text, pdu.Id(), DeliverSm)
				}

				if dcs, err := pdu.GetMainAsUint32(DataCoding); err != nil {
					t.Fatalf("failed get DataCoding: %v", err)
				} else {
					if dcs != test.expected.dcs {
						t.Fatalf("with text %v, dcs %v not equals expected %v", test.text, dcs, test.expected.dcs)
					}
				}

				if esmClass, err := pdu.GetMainAsUint32(ESMClass); err != nil {
					t.Fatalf("failed get EsmClass: %v", err)
				} else {
					if esmClass != test.expected.esmClass {
						t.Fatalf("with text %v, esmClass %v not equals expected %v", test.text, esmClass, test.expected.esmClass)
					}
				}

				if len(pdus) == 1 {
					continue
				}

				if sm, err := pdu.GetMainAsRaw(ShortMessage); err != nil {
					t.Fatalf("failed get ShortMessage: %v", err)
				} else {
					if len(sm) < 7 {
						t.Fatalf("bad sm length %v", len(sm))
					}

					if int(sm[5]) != test.expected.totalParts {
						t.Fatalf("with text %v, totalParts %v not equals expected %v", test.text, sm[5], test.expected.totalParts)
					}

					if int(sm[6]) != i + 1 {
						t.Fatalf("with text %v, part %v not equals expected %v", test.text, sm[6], i)
					}

					if v := uint8(test.msgRefNumProvider() >> 8); sm[3] != v {
						t.Fatalf("with text %v, first byte msg ref num %v not equals expected %v", test.text, sm[3], v)
					}

					if v := uint8(test.msgRefNumProvider()); sm[4] != v {
						t.Fatalf("with text %v, second byte msg ref num %v not equals expected %v", test.text, sm[4], v)
					}
				}
			}
		}
	}
}
