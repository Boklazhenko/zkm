package zkm

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
	id              Id
	status          Status
	seq             uint32
	mandatoryParams *mandatoryParams
	optionalParams  *optionalParams
	raw             []byte
}

func NewEmptyPdu() *Pdu {
	return &Pdu{
		id:              0,
		mandatoryParams: newMandatoryParams(0),
		optionalParams:  newOptionalParams(),
		raw:             nil,
	}
}

func NewPdu(id Id) *Pdu {
	return &Pdu{
		id:              id,
		mandatoryParams: newMandatoryParams(id),
		optionalParams:  newOptionalParams(),
		raw:             nil,
	}
}

func (pdu *Pdu) Id() Id {
	return pdu.id
}

func (pdu *Pdu) Status() Status {
	return pdu.status
}

func (pdu *Pdu) SetStatus(status Status) {
	pdu.status = status
	if pdu.raw != nil {
		rawStatus := make([]byte, pduHeaderPartSize)
		binary.BigEndian.PutUint32(rawStatus, uint32(status))
		copy(pdu.raw[2*pduHeaderPartSize:], rawStatus)
	}
}

func (pdu *Pdu) Seq() uint32 {
	return pdu.seq
}

func (pdu *Pdu) SetSeq(seq uint32) {
	pdu.seq = seq
	if pdu.raw != nil {
		rawSeq := make([]byte, pduHeaderPartSize)
		binary.BigEndian.PutUint32(rawSeq, seq)
		copy(pdu.raw[3*pduHeaderPartSize:], rawSeq)
	}
}

func (pdu *Pdu) String() string {
	return fmt.Sprintf("id:%v status:%v seq:%v", pdu.Id, pdu.Status, pdu.Seq)
}

func (pdu *Pdu) IsReq() bool {
	return pdu.id&0x80000000 == 0
}

func (pdu *Pdu) CreateResp(status Status) (*Pdu, error) {
	var resp *Pdu
	switch pdu.id {
	case BindReceiver:
		resp = NewPdu(BindReceiverResp)
	case BindTransmitter:
		resp = NewPdu(BindTransmitterResp)
	case QuerySm:
		resp = NewPdu(QuerySmResp)
	case SubmitSm:
		resp = NewPdu(SubmitSmResp)
	case DeliverSm:
		resp = NewPdu(DeliverSmResp)
	case Unbind:
		resp = NewPdu(UnbindResp)
	case ReplaceSm:
		resp = NewPdu(ReplaceSmResp)
	case CancelSm:
		resp = NewPdu(CancelSmResp)
	case BindTransceiver:
		resp = NewPdu(BindTransceiverResp)
	case EnquireLink:
		resp = NewPdu(EnquireLinkResp)
	case SubmitMulti:
		resp = NewPdu(SubmitMultiResp)
	case DataSm:
		resp = NewPdu(DataSmResp)
	default:
		return nil, fmt.Errorf("cann't create resp for cmd id [%v]", pdu.Id)
	}

	resp.SetSeq(pdu.Seq())
	resp.SetStatus(status)
	return resp, nil
}

func (pdu *Pdu) Serialize() []byte {
	if pdu.raw != nil {
		return pdu.raw
	}

	buff := bytes.Buffer{}
	b := make([]byte, pduHeaderPartSize)
	binary.BigEndian.PutUint32(b, pdu.Len())
	buff.Write(b)
	binary.BigEndian.PutUint32(b, uint32(pdu.id))
	buff.Write(b)
	binary.BigEndian.PutUint32(b, uint32(pdu.status))
	buff.Write(b)
	binary.BigEndian.PutUint32(b, pdu.seq)
	buff.Write(b)
	buff.Write(pdu.mandatoryParams.serialize())
	buff.Write(pdu.optionalParams.serialize())

	pdu.raw = buff.Bytes()
	return pdu.raw
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
	pdu.id = Id(binary.BigEndian.Uint32(b))

	n, err = buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	pdu.status = Status(binary.BigEndian.Uint32(b))

	n, err = buff.Read(b)
	if n != pduHeaderPartSize {
		return err
	}
	pdu.seq = binary.BigEndian.Uint32(b)

	pdu.mandatoryParams = newMandatoryParams(pdu.id)

	if err = pdu.mandatoryParams.deserialize(buff); err != nil {
		return err
	}

	pdu.optionalParams = newOptionalParams()

	if err = pdu.optionalParams.deserialize(buff); err != nil {
		return err
	}

	pdu.raw = raw
	return nil
}

func (pdu *Pdu) SetMain(name Name, value interface{}) error {
	p, err := pdu.mandatoryParams.get(name)

	if err != nil {
		return err
	}

	pdu.raw = nil

	return p.value().set(value)
}

func (pdu *Pdu) GetMainAsRaw(name Name) ([]byte, error) {
	p, err := pdu.mandatoryParams.get(name)

	if err != nil {
		return nil, err
	}

	return p.value().raw(), nil
}

func (pdu *Pdu) GetMainAsString(name Name) (string, error) {
	p, err := pdu.mandatoryParams.get(name)

	if err != nil {
		return "", err
	}

	return p.value().String(), nil
}

func (pdu *Pdu) GetMainAsUint32(name Name) (uint32, error) {
	p, err := pdu.mandatoryParams.get(name)

	if err != nil {
		return 0, err
	}

	return p.value().uint32()
}

func (pdu *Pdu) SetOpt(tag Tag, value interface{}) error {
	pdu.raw = nil
	return pdu.optionalParams.add(tag, value)
}

func (pdu *Pdu) RemoveOpt(tag Tag) {
	pdu.optionalParams.remove(tag)
}

func (pdu *Pdu) GetOptAsRaw(tag Tag) ([]byte, error) {
	p, err := pdu.optionalParams.get(tag)

	if err != nil {
		return nil, err
	}

	return p.value().raw(), nil
}

func (pdu *Pdu) GetOptAsString(tag Tag) (string, error) {
	p, err := pdu.optionalParams.get(tag)

	if err != nil {
		return "", err
	}

	return p.value().String(), nil
}

func (pdu *Pdu) GetOptAsUint32(tag Tag) (uint32, error) {
	p, err := pdu.optionalParams.get(tag)

	if err != nil {
		return 0, err
	}

	return p.value().uint32()
}

func (pdu *Pdu) Len() uint32 {
	return 4*pduHeaderPartSize + pdu.mandatoryParams.len() + pdu.optionalParams.len()
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
