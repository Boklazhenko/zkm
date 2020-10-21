package pdu

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Id uint32
type Status uint32

const pduHeaderPartSize = 4

const (
	GenericNack         Id = 0x80000000
	BindReceiver        Id = 0x00000001
	BindReceiverResp    Id = 0x80000001
	BindTransmitter     Id = 0x00000002
	BindTransmitterResp Id = 0x80000002
	QuerySm             Id = 0x00000003
	QuerySmResp         Id = 0x80000003
	SubmitSm            Id = 0x00000004
	SubmitSmResp        Id = 0x80000004
	DeliverSm           Id = 0x00000005
	DeliverSmResp       Id = 0x80000005
	Unbind              Id = 0x00000006
	UnbindResp          Id = 0x80000006
	ReplaceSm           Id = 0x00000007
	ReplaceSmResp       Id = 0x80000007
	CancelSm            Id = 0x00000008
	CancelSmResp        Id = 0x80000008
	BindTransceiver     Id = 0x00000009
	BindTransceiverResp Id = 0x80000009
	Outbind             Id = 0x0000000B
	EnquireLink         Id = 0x00000015
	EnquireLinkResp     Id = 0x80000015
	SubmitMulti         Id = 0x00000021
	SubmitMultiResp     Id = 0x80000021
	AlertNotification   Id = 0x00000102
	DataSm              Id = 0x00000103
	DataSmResp          Id = 0x80000103
)

const (
	EsmeROk              Status = 0x00000000
	EsmeRInvMsgLen       Status = 0x00000001
	EsmeRInvCmdLen       Status = 0x00000002
	EsmeRInvCmdId        Status = 0x00000003
	EsmeRInvBndSts       Status = 0x00000004
	EsmeRAlyBnd          Status = 0x00000005
	EsmeRInvPrtFlg       Status = 0x00000006
	EsmeRInvRegDlvFlg    Status = 0x00000007
	EsmeRSysErr          Status = 0x00000008
	EsmeRInvSrcAdr       Status = 0x0000000a
	EsmeRInvDstAdr       Status = 0x0000000b
	EsmeRInvMsgId        Status = 0x0000000c
	EsmeRBindFail        Status = 0x0000000d
	EsmeRInvPaswd        Status = 0x0000000e
	EsmeRInvSysId        Status = 0x0000000f
	EsmeRCancelFail      Status = 0x00000011
	EsmeRReplaceFail     Status = 0x00000013
	EsmeRMsgQFul         Status = 0x00000014
	EsmeRInvSerTyp       Status = 0x00000015
	EsmeRInvNumDests     Status = 0x00000033
	EsmeRInvDLName       Status = 0x00000034
	EsmeRInvDestFlag     Status = 0x00000040
	EsmeRInvSubRep       Status = 0x00000042
	EsmeRInvEsmClass     Status = 0x00000043
	EsmeRCntSubDL        Status = 0x00000044
	EsmeRSubmitFail      Status = 0x00000045
	EsmeRInvSrcTon       Status = 0x00000048
	EsmeRInvSrcNpi       Status = 0x00000049
	EsmeRInvDstTon       Status = 0x00000050
	EsmeRInvDstNpi       Status = 0x00000051
	EsmeRInvSysTyp       Status = 0x00000053
	EsmeRInvRepFlag      Status = 0x00000054
	EsmeRInvNumMsgs      Status = 0x00000055
	EsmeRThrottled       Status = 0x00000058
	EsmeRInvSched        Status = 0x00000061
	EsmeRInvExpire       Status = 0x00000062
	EsmeRInvDftMsgId     Status = 0x00000063
	EsmeRxTAppn          Status = 0x00000064
	EsmeRxPAppn          Status = 0x00000065
	EsmeRxRAppn          Status = 0x00000066
	EsmeRQueryFail       Status = 0x00000067
	EsmeRInvOptParStream Status = 0x000000c0
	EsmeROptParNotAllwd  Status = 0x000000c1
	EsmeRInvParLen       Status = 0x000000c2
	EsmeRMissingOptParam Status = 0x000000c3
	EsmeRInvOptParamVal  Status = 0x000000c4
	EsmeRDeliveryFailure Status = 0x000000fe
	EsmeRUnknownErr      Status = 0x000000ff
)

type Pdu struct {
	Id              Id
	Status          Status
	Seq             uint32
	MandatoryParams *MandatoryParams
	OptionalParams  *OptionalParams
}

func New(id Id) *Pdu {
	return &Pdu{
		Id:              id,
		MandatoryParams: NewMandatoryParams(id),
		OptionalParams:  NewOptionalParams(),
	}
}

func (pdu *Pdu) Serialize() []byte {
	buff := bytes.Buffer{}
	b := make([]byte, pduHeaderPartSize)
	binary.BigEndian.PutUint32(b, pdu.Len())
	buff.Write(b)
	binary.BigEndian.PutUint32(b, uint32(pdu.Id))
	buff.Write(b)
	binary.BigEndian.PutUint32(b, uint32(pdu.Status))
	buff.Write(b)
	binary.BigEndian.PutUint32(b, pdu.Seq)
	buff.Write(b)
	buff.Write(pdu.MandatoryParams.Serialize())
	buff.Write(pdu.OptionalParams.Serialize())
	return buff.Bytes()
}

func (pdu *Pdu) Deserialize(raw []byte) error {
	buff := bytes.NewBuffer(raw)

	b := make([]byte, pduHeaderPartSize)

	n, err := buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	l := binary.BigEndian.Uint32(b)

	if int(l) != len(raw) {
		return fmt.Errorf("bad pdu length: in field - %v, received - %v", l, len(raw))
	}

	n, err = buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	pdu.Id = Id(binary.BigEndian.Uint32(b))

	n, err = buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	pdu.Status = Status(binary.BigEndian.Uint32(b))

	n, err = buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	pdu.Seq = binary.BigEndian.Uint32(b)

	pdu.MandatoryParams = NewMandatoryParams(pdu.Id)

	if err = pdu.MandatoryParams.Deserialize(buff); err != nil {
		return err
	}

	pdu.OptionalParams = NewOptionalParams()

	return pdu.OptionalParams.Deserialize(buff)
}

func (pdu *Pdu) Len() uint32 {
	return 4*pduHeaderPartSize + pdu.MandatoryParams.Len() + pdu.OptionalParams.Len()
}

func (id Id) String() string {
	switch id {
	case GenericNack:
		return "GenericNack"
	case BindReceiver:
		return "BindReceiver"
	case BindReceiverResp:
		return "BindReceiverResp"
	case BindTransmitter:
		return "BindTransmitter"
	case BindTransmitterResp:
		return "BindTransmitterResp"
	case QuerySm:
		return "QuerySm"
	case QuerySmResp:
		return "QuerySmResp"
	case SubmitSm:
		return "SubmitSm"
	case SubmitSmResp:
		return "SubmitSmResp"
	case DeliverSm:
		return "DeliverSm"
	case DeliverSmResp:
		return "DeliverSmResp"
	case Unbind:
		return "Unbind"
	case UnbindResp:
		return "UnbindResp"
	case ReplaceSm:
		return "ReplaceSm"
	case ReplaceSmResp:
		return "ReplaceSmResp"
	case CancelSm:
		return "CancelSm"
	case CancelSmResp:
		return "CancelSmResp"
	case BindTransceiver:
		return "BindTransceiver"
	case BindTransceiverResp:
		return "BindTransceiverResp"
	case Outbind:
		return "Outbind"
	case EnquireLink:
		return "EnquireLink"
	case EnquireLinkResp:
		return "EnquireLinkResp"
	case SubmitMulti:
		return "SubmitMulti"
	case SubmitMultiResp:
		return "SubmitMultiResp"
	case AlertNotification:
		return "AlertNotification"
	case DataSm:
		return "DataSm"
	case DataSmResp:
		return "DataSmResp"
	default:
		return "Unknown"
	}
}

func (s Status) String() string {
	switch s {
	case EsmeROk:
		return "ok"
	case EsmeRInvMsgLen:
		return "invalid message length"
	case EsmeRInvCmdLen:
		return "invalid command length"
	case EsmeRInvCmdId:
		return "invalid command id"
	case EsmeRInvBndSts:
		return "incorrect bind status for given command"
	case EsmeRAlyBnd:
		return "already in bound state"
	case EsmeRInvPrtFlg:
		return "invalid priority flag"
	case EsmeRInvRegDlvFlg:
		return "invalid registered delivery flag"
	case EsmeRSysErr:
		return "system error"
	case EsmeRInvSrcAdr:
		return "invalid source address"
	case EsmeRInvDstAdr:
		return "invalid destination address"
	case EsmeRInvMsgId:
		return "invalid message id"
	case EsmeRBindFail:
		return "bind failed"
	case EsmeRInvPaswd:
		return "invalid password"
	case EsmeRInvSysId:
		return "invalid system id"
	case EsmeRCancelFail:
		return "cancelsm failed"
	case EsmeRReplaceFail:
		return "replacesm failed"
	case EsmeRMsgQFul:
		return "message queue full"
	case EsmeRInvSerTyp:
		return "invalid service type"
	case EsmeRInvNumDests:
		return "invalid number of destinations"
	case EsmeRInvDLName:
		return "invalid distribution list name"
	case EsmeRInvDestFlag:
		return "invalid destination flag"
	case EsmeRInvSubRep:
		return "invalid 'submit with replace' request"
	case EsmeRInvEsmClass:
		return "invalid esm class field data"
	case EsmeRCntSubDL:
		return "cannot submit to distribution list"
	case EsmeRSubmitFail:
		return "submitsm or submitmulti failed"
	case EsmeRInvSrcTon:
		return "invalid source address ton"
	case EsmeRInvSrcNpi:
		return "invalid source address npi"
	case EsmeRInvDstTon:
		return "invalid destination address ton"
	case EsmeRInvDstNpi:
		return "invalid destination address npi"
	case EsmeRInvSysTyp:
		return "invalid system type field"
	case EsmeRInvRepFlag:
		return "invalid replace_if_present flag"
	case EsmeRInvNumMsgs:
		return "invalid number of messages"
	case EsmeRThrottled:
		return "throttling error"
	case EsmeRInvSched:
		return "invalid scheduled delivery time"
	case EsmeRInvExpire:
		return "invalid message validity period (expiry time)"
	case EsmeRInvDftMsgId:
		return "predefined message invalid or not found"
	case EsmeRxTAppn:
		return "esme receiver temporary app error code"
	case EsmeRxPAppn:
		return "esme receiver permanent app error code"
	case EsmeRxRAppn:
		return "esme receiver reject message error code"
	case EsmeRQueryFail:
		return "querysm request failed"
	case EsmeRInvOptParStream:
		return "error in the optional part of the pdu body"
	case EsmeROptParNotAllwd:
		return "optional parameter not allowed"
	case EsmeRInvParLen:
		return "invalid parameter length"
	case EsmeRMissingOptParam:
		return "expected optional parameter missing"
	case EsmeRInvOptParamVal:
		return "invalid optional parameter value"
	case EsmeRDeliveryFailure:
		return "delivery failure (used for datasmresp)"
	case EsmeRUnknownErr:
		return "unknown error"
	default:
		return "unknown"
	}
}
