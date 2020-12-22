package zkm

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
	Text  string
}

//"id:cb9c40f1-0aa1-4b3e-afb8-7dd24d03a716 sub:001 dlvrd:001 submit date:2010241205 done date:2010241206 stat:DELIVRD err:000"

func NewDeliveryReceiptInfoByPdu(pdu *Pdu) *DeliveryReceiptInfo {
	dri := &DeliveryReceiptInfo{}

	if pdu.id != DeliverSm {
		return dri
	}

	if esmClass, err := pdu.GetMainAsUint32(ESMClass); (esmClass&0x24) == 0 || err != nil {
		return dri
	}

	rawText, err := pdu.GetMainAsRaw(ShortMessage)
	if err != nil {
		return dri
	}

	dri.Text = string(rawText)

	rawFields := strings.Split(dri.Text, " ")

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

func NewDeliveryReceiptInfo(msgId string, submitTime, doneTime int64, state DeliveryReceiptState, err uint16) *DeliveryReceiptInfo {
	success := 1
	if state != Delivered {
		success = 0
	}

	return &DeliveryReceiptInfo{
		Id:    msgId,
		State: state,
		Err:   err,
		Text: fmt.Sprintf("id:%v sub:001 dlvrd:00%v submit date:%v done date:%v stat:%v err:%v",
			msgId, success, time.Unix(submitTime, 0).Format("0601021504"),
			time.Unix(doneTime, 0).Format("0601021504"),
			state, err%1000),
	}
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
