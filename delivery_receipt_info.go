package zkm

import (
	"strconv"
	"strings"
)

type DeliveryReceiptState uint8

const (
	EnRoute DeliveryReceiptState = iota + 1
	Delivered
	Expired
	Deleted
	Undeliverable
	Accepted
	Unknown
	Rejected
)

type DeliveryReceiptInfo struct {
	Id    string
	State DeliveryReceiptState
	Err   uint16
}

//"id:cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716 sub:001 dlvrd:001 submit " +
//					"date:2010241205 done date:2010241206 stat:DELIVRD err:000"

func NewDeliveryReceiptInfo(pdu *Pdu) *DeliveryReceiptInfo {
	dri := &DeliveryReceiptInfo{}

	if pdu.Id != DeliverSm {
		return dri
	}

	if esmClass, err := pdu.GetMainAsUint32(ESMClass); (esmClass&0x24) == 0 || err != nil {
		return dri
	}

	rawText, err := pdu.GetMainAsRaw(ShortMessage)
	if err != nil {
		return dri
	}

	rawFields := strings.Split(string(rawText), " ")

	for _, rawF := range rawFields {
		f := strings.Split(rawF, ":")

		if len(f) != 2 {
			continue
		}

		switch f[0] {
		case "id":
			dri.Id = f[1]
		case "stat":
			switch f[1] {
			case "ENROUTE":
				dri.State = EnRoute
			case "DELIVRD":
				dri.State = Delivered
			case "EXPIRED":
				dri.State = Expired
			case "DELETED":
				dri.State = Deleted
			case "UNDELIV":
				dri.State = Undeliverable
			case "ACCEPTD":
				dri.State = Accepted
			case "UNKNOWN":
				dri.State = Unknown
			case "REJECTD":
				dri.State = Rejected
			}
		case "err":
			e, err := strconv.ParseUint(f[1], 10, 64)
			if err == nil {
				dri.Err = uint16(e)
			}
		}
	}

	return dri
}

func (s DeliveryReceiptState) String() string {
	switch s {
	case EnRoute:
		return "ENROUTE"
	case Delivered:
		return "DELIVRD"
	case Expired:
		return "EXPIRED"
	case Deleted:
		return "DELETED"
	case Undeliverable:
		return "UNDELIV"
	case Accepted:
		return "ACCEPTD"
	case Unknown:
		return "UNKNOWN"
	case Rejected:
		return "REJECTD"
	default:
		return "UNKNOWN"
	}
}
