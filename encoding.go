package zkm

import (
	encoding2 "github.com/Boklazhenko/zkm/internal/encoding"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type DataCodingScheme uint8

const (
	SmscDefaultAlphabetScheme = 0x00
	AsciiScheme               = 0x01
	Latin1Scheme              = 0x03
	Ucs2Scheme                = 0x08
)

func Gsm7Packed() encoding.Encoding {
	return encoding2.GSM7(true)
}

func Gsm7Unpacked() encoding.Encoding {
	return encoding2.GSM7(false)
}

func Latin1() encoding.Encoding {
	return charmap.ISO8859_1
}

func Ucs2() encoding.Encoding {
	return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
}

func Encode(s string, e encoding.Encoding) ([]byte, error) {
	encoder := e.NewEncoder()
	r, _, err := transform.Bytes(encoder, []byte(s))
	return r, err
}

func Decode(b []byte, e encoding.Encoding) (string, error) {
	decoder := e.NewDecoder()
	r, _, err := transform.Bytes(decoder, b)
	return string(r), err
}
