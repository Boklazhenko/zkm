package pdu

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

var paramNotFound = fmt.Errorf("not found")
var paramBadType = fmt.Errorf("bad type")

type Name string

const (
	AddrNPI              Name = "addr_npi"
	AddrTON              Name = "addr_ton"
	AddressRange         Name = "address_range"
	DataCoding           Name = "data_coding"
	DestAddrNPI          Name = "dest_addr_npi"
	DestAddrTON          Name = "dest_addr_ton"
	DestinationAddr      Name = "destination_addr"
	ESMClass             Name = "esm_class"
	ErrorCode            Name = "error_code"
	FinalDate            Name = "final_date"
	InterfaceVersion     Name = "interface_version"
	MessageID            Name = "message_id"
	MessageState         Name = "message_state"
	NumberDests          Name = "number_of_dests"
	NoUnsuccess          Name = "no_unsuccess"
	Password             Name = "password"
	PriorityFlag         Name = "priority_flag"
	ProtocolID           Name = "protocol_id"
	RegisteredDelivery   Name = "registered_delivery"
	ReplaceIfPresentFlag Name = "replace_if_present_flag"
	SMDefaultMsgID       Name = "sm_default_msg_id"
	SMLength             Name = "sm_length"
	ScheduleDeliveryTime Name = "schedule_delivery_time"
	ServiceType          Name = "service_type"
	ShortMessage         Name = "short_message"
	SourceAddr           Name = "source_addr"
	SourceAddrNPI        Name = "source_addr_npi"
	SourceAddrTON        Name = "source_addr_ton"
	SystemID             Name = "system_id"
	SystemType           Name = "system_type"
	ValidityPeriod       Name = "validity_period"
	EsmeAddr             Name = "esme_addr"
	DlName               Name = "dl_name"
	EsmeAddrTON          Name = "esme_addr_ton"
	EsmeAddrNPI          Name = "esme_addr_npi"
	DestFlag             Name = "dest_flag"
)

type Tag uint16

const (
	TagDestAddrSubunit          Tag = 0x0005
	TagDestNetworkType          Tag = 0x0006
	TagDestBearerType           Tag = 0x0007
	TagDestTelematicsID         Tag = 0x0008
	TagSourceAddrSubunit        Tag = 0x000D
	TagSourceNetworkType        Tag = 0x000E
	TagSourceBearerType         Tag = 0x000F
	TagSourceTelematicsID       Tag = 0x0010
	TagQosTimeToLive            Tag = 0x0017
	TagPayloadType              Tag = 0x0019
	TagAdditionalStatusInfoText Tag = 0x001D
	TagReceiptedMessageID       Tag = 0x001E
	TagMsMsgWaitFacilities      Tag = 0x0030
	TagPrivacyIndicator         Tag = 0x0201
	TagSourceSubaddress         Tag = 0x0202
	TagDestSubaddress           Tag = 0x0203
	TagUserMessageReference     Tag = 0x0204
	TagUserResponseCode         Tag = 0x0205
	TagSourcePort               Tag = 0x020A
	TagDestinationPort          Tag = 0x020B
	TagSarMsgRefNum             Tag = 0x020C
	TagLanguageIndicator        Tag = 0x020D
	TagSarTotalSegments         Tag = 0x020E
	TagSarSegmentSeqnum         Tag = 0x020F
	TagScInterfaceVersion       Tag = 0x0210
	TagCallbackNumPresInd       Tag = 0x0302
	TagCallbackNumAtag          Tag = 0x0303
	TagNumberOfMessages         Tag = 0x0304
	TagCallbackNum              Tag = 0x0381
	TagDpfResult                Tag = 0x0420
	TagSetDpf                   Tag = 0x0421
	TagMsAvailabilityStatus     Tag = 0x0422
	TagNetworkErrorCode         Tag = 0x0423
	TagMessagePayload           Tag = 0x0424
	TagDeliveryFailureReason    Tag = 0x0425
	TagMoreMessagesToSend       Tag = 0x0426
	TagMessageStateOption       Tag = 0x0427
	TagUssdServiceOp            Tag = 0x0501
	TagDisplayTime              Tag = 0x1201
	TagSmsSignal                Tag = 0x1203
	TagMsValidity               Tag = 0x1204
	TagAlertOnMessageDelivery   Tag = 0x130C
	TagItsReplyType             Tag = 0x1380
	TagItsSessionInfo           Tag = 0x1383
)

type MandatoryParams struct {
	names  []Name
	params map[Name]*MandatoryParam
}

func NewMandatoryParams(id Id) *MandatoryParams {
	names := make([]Name, 0)
	switch id {
	case BindReceiver, BindTransceiver, BindTransmitter:
		names = []Name{
			SystemID,
			Password,
			SystemType,
			InterfaceVersion,
			AddrTON,
			AddrNPI,
			AddressRange,
		}
	case BindReceiverResp, BindTransceiverResp, BindTransmitterResp:
		names = []Name{
			SystemID,
		}
	case SubmitSm, DeliverSm:
		names = []Name{
			ServiceType,
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
			DestAddrTON,
			DestAddrNPI,
			DestinationAddr,
			ESMClass,
			ProtocolID,
			PriorityFlag,
			ScheduleDeliveryTime,
			ValidityPeriod,
			RegisteredDelivery,
			ReplaceIfPresentFlag,
			DataCoding,
			SMDefaultMsgID,
			SMLength,
			ShortMessage,
		}
	case SubmitSmResp, DeliverSmResp, DataSmResp:
		names = []Name{
			MessageID,
		}
	case EnquireLink, EnquireLinkResp, GenericNack, Unbind, UnbindResp, CancelSmResp, ReplaceSmResp:
	case Outbind:
		names = []Name{
			SystemID,
			Password,
		}
	case DataSm:
		names = []Name{
			ServiceType,
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
			DestAddrTON,
			DestAddrNPI,
			DestinationAddr,
			ESMClass,
			RegisteredDelivery,
			DataCoding,
		}
	case QuerySm:
		names = []Name{
			MessageID,
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
		}
	case QuerySmResp:
		names = []Name{
			MessageID,
			FinalDate,
			MessageState,
			ErrorCode,
		}
	case CancelSm:
		names = []Name{
			ServiceType,
			MessageID,
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
			DestAddrTON,
			DestAddrNPI,
			DestinationAddr,
		}
	case ReplaceSm:
		names = []Name{
			MessageID,
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
			ScheduleDeliveryTime,
			ValidityPeriod,
			RegisteredDelivery,
			SMDefaultMsgID,
			SMLength,
			ShortMessage,
		}
	case AlertNotification:
		names = []Name{
			SourceAddrTON,
			SourceAddrNPI,
			SourceAddr,
			EsmeAddrTON,
			EsmeAddrNPI,
			EsmeAddr,
		}
	}

	params := make(map[Name]*MandatoryParam)
	for _, n := range names {
		params[n] = NewMandatoryParam(n)
	}

	return &MandatoryParams{names: names, params: params}
}

func (ps *MandatoryParams) Len() uint32 {
	totalLen := 0
	for _, p := range ps.params {
		totalLen += p.Value().Len()
	}

	return uint32(totalLen)
}

func (ps *MandatoryParams) Serialize() []byte {
	buff := bytes.Buffer{}
	for _, n := range ps.names {
		buff.Write(ps.params[n].Value().Raw())
	}
	return buff.Bytes()
}

func (ps *MandatoryParams) Deserialize(buff *bytes.Buffer) error {
	for _, n := range ps.names {
		if err := ps.params[n].Value().Deserialize(buff); err != nil {
			return err
		}

		if n == SMLength {
			smLength, err := ps.params[n].Value().Uint32()
			if err != nil {
				panic(err)
			}
			ps.params[ShortMessage].value = NewOctetStringValue(int(smLength))
		}
	}
	return nil
}

func (ps *MandatoryParams) Get(name Name) (*MandatoryParam, error) {
	if p, ok := ps.params[name]; ok {
		return p, nil
	} else {
		return nil, paramNotFound
	}
}

type MandatoryParam struct {
	name  Name
	value Value
}

func NewMandatoryParam(name Name) *MandatoryParam {
	var value Value
	switch name {
	case InterfaceVersion, AddrTON, AddrNPI, SourceAddrTON, SourceAddrNPI, DestAddrTON, DestAddrNPI, EsmeAddrTON,
		EsmeAddrNPI, ESMClass, ProtocolID, PriorityFlag, RegisteredDelivery, ReplaceIfPresentFlag, DataCoding,
		SMDefaultMsgID, SMLength, NumberDests, DestFlag, NoUnsuccess, MessageState, ErrorCode:
		value = NewUint8Value()
	case ScheduleDeliveryTime, ValidityPeriod, FinalDate:
		value = NewFixedCOctetStringValue(17)
	case SourceAddr, DestinationAddr, DlName:
		value = NewCOctetStringValue(21)
	case ShortMessage:
		value = NewOctetStringValue(0)
	case SystemID:
		value = NewCOctetStringValue(16)
	case Password:
		value = NewCOctetStringValue(9)
	case SystemType:
		value = NewCOctetStringValue(13)
	case AddressRange:
		value = NewCOctetStringValue(41)
	case MessageID, EsmeAddr:
		value = NewCOctetStringValue(65)
	case ServiceType:
		value = NewCOctetStringValue(6)
	default:
		value = NewOctetStringValue(0)
	}

	return &MandatoryParam{name: name, value: value}
}

func (p *MandatoryParam) String() string {
	return fmt.Sprintf("%v: %v", p.name, p.value)
}

func (p *MandatoryParam) Name() Name {
	return p.name
}

func (p *MandatoryParam) Value() Value {
	return p.value
}

type OptionalParam struct {
	tag   Tag
	value Value
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	} else {
		return b
	}
}

type OptionalParams struct {
	params map[Tag]*OptionalParam
}

func NewOptionalParams() *OptionalParams {
	return &OptionalParams{params: make(map[Tag]*OptionalParam)}
}

func (ps *OptionalParams) Len() uint32 {
	totalLen := 0
	for _, p := range ps.params {
		totalLen += 4 + p.Value().Len()
	}

	return uint32(totalLen)
}

func (ps *OptionalParams) Serialize() []byte {
	buff := bytes.Buffer{}
	for _, p := range ps.params {
		buff.Write(p.Value().Raw())
	}
	return buff.Bytes()
}

func (ps *OptionalParams) Deserialize(buff *bytes.Buffer) error {
	for buff.Len() > 0 {
		b := make([]byte, 4)
		n, err := buff.Read(b)
		if n != len(b) {
			return fmt.Errorf("not enough data")
		}

		if err != nil {
			return err
		}

		tag := Tag(binary.BigEndian.Uint16(b[:2]))
		l := binary.BigEndian.Uint16(b[2:])
		optParam := NewOptionalParam(tag, l)

		if err = optParam.Value().Deserialize(buff); err != nil {
			return err
		}

		if l != uint16(optParam.Value().Len()) {
			return fmt.Errorf("bad optional param: len %v, real len %v", l, optParam.Value().Len())
		}

		ps.params[tag] = optParam
	}

	return nil
}

func (ps *OptionalParams) Get(tag Tag) (*OptionalParam, error) {
	if p, ok := ps.params[tag]; ok {
		return p, nil
	} else {
		return nil, paramNotFound
	}
}

func (ps *OptionalParams) Add(tag Tag, d interface{}) error {
	p := NewOptionalParam(tag, 0)
	if err := p.Value().Set(d); err != nil {
		return err
	}
	ps.params[tag] = p
	return nil
}

func NewOptionalParam(tag Tag, len uint16) *OptionalParam {
	var value Value
	switch tag {
	case TagDestAddrSubunit, TagDestNetworkType, TagSourceNetworkType, TagDestBearerType, TagSourceBearerType,
		TagSourceAddrSubunit, TagSourceTelematicsID, TagPayloadType, TagMsMsgWaitFacilities, TagPrivacyIndicator,
		TagUserResponseCode, TagLanguageIndicator, TagSarTotalSegments, TagSarSegmentSeqnum, TagScInterfaceVersion,
		TagDisplayTime, TagMsValidity, TagDpfResult, TagSetDpf, TagMsAvailabilityStatus, TagDeliveryFailureReason,
		TagMoreMessagesToSend, TagMessageStateOption, TagCallbackNumPresInd, TagNumberOfMessages, TagItsReplyType, TagUssdServiceOp:
		value = NewUint8Value()
	case TagDestTelematicsID, TagUserMessageReference, TagSourcePort, TagDestinationPort,
		TagSarMsgRefNum, TagSmsSignal, TagItsSessionInfo:
		value = NewUint16Value()
	case TagQosTimeToLive:
		value = NewUint32Value()
	case TagAdditionalStatusInfoText:
		value = NewCOctetStringValue(256)
	case TagReceiptedMessageID:
		value = NewCOctetStringValue(65)
	case TagSourceSubaddress, TagDestSubaddress:
		value = NewOctetStringValue(int(min(23, len)))
	case TagNetworkErrorCode:
		value = NewOctetStringValue(3)
	case TagMessagePayload:
		value = NewOctetStringValue(int(min(math.MaxUint16, len)))
	case TagCallbackNum:
		value = NewOctetStringValue(int(min(19, len)))
	default:
		value = NewOctetStringValue(0)
	}

	return &OptionalParam{tag: tag, value: value}
}

func (p *OptionalParam) String() string {
	return fmt.Sprintf("%v: %v", p.tag, p.value)
}

func (p *OptionalParam) Tag() Tag {
	return p.tag
}

func (p *OptionalParam) Value() Value {
	return p.value
}

type Value interface {
	fmt.Stringer
	Len() int
	Set(d interface{}) error
	Deserialize(buff *bytes.Buffer) error
	Raw() []byte
	Uint32() (uint32, error)
}

type OctetStringValue struct {
	raw []byte
}

func NewOctetStringValue(len int) *OctetStringValue {
	return &OctetStringValue{raw: make([]byte, len)}
}

func (v *OctetStringValue) String() string {
	return fmt.Sprintf("%v", v.raw)
}

func (v *OctetStringValue) Len() int {
	return len(v.raw)
}

func (v *OctetStringValue) Set(d interface{}) error {
	switch data := d.(type) {
	case []byte:
		v.raw = data
	default:
		return paramBadType
	}

	return nil
}

func (v *OctetStringValue) Deserialize(buff *bytes.Buffer) error {
	bytes := make([]byte, v.Len())
	if n, _ := buff.Read(bytes); n == v.Len() {
		v.raw = bytes
		return nil
	} else {
		return fmt.Errorf("readed %v bytes, need %v", n, len(v.raw))
	}
}

func (v *OctetStringValue) Raw() []byte {
	return v.raw
}

func (v *OctetStringValue) Uint32() (uint32, error) {
	return 0, paramBadType
}

type COctetStringValue struct {
	*OctetStringValue
	maxLen int
}

func (v *COctetStringValue) String() string {
	return string(v.raw[:len(v.raw)-1])
}

func (v *COctetStringValue) Set(d interface{}) error {
	switch data := d.(type) {
	case string:
		if err := v.Deserialize(bytes.NewBuffer(append([]byte(data), 0))); err != nil {
			return err
		}
	default:
		return paramBadType
	}

	return nil
}

func (v *COctetStringValue) Deserialize(buff *bytes.Buffer) error {
	if bytes, err := buff.ReadBytes(0x00); err != nil {
		return err
	} else if len(bytes) > v.maxLen {
		return fmt.Errorf("real length %v exceeded the maximum length %v", len(v.raw), v.maxLen)
	} else {
		v.raw = bytes
		return nil
	}
}

func NewCOctetStringValue(maxLen int) *COctetStringValue {
	return &COctetStringValue{OctetStringValue: NewOctetStringValue(1), maxLen: maxLen}
}

type FixedCOctetStringValue struct {
	*COctetStringValue
}

func NewFixedCOctetStringValue(len int) *FixedCOctetStringValue {
	return &FixedCOctetStringValue{COctetStringValue: NewCOctetStringValue(len)}
}

func (v *FixedCOctetStringValue) Set(d interface{}) error {
	switch data := d.(type) {
	case string:
		if err := v.Deserialize(bytes.NewBuffer(append([]byte(data), 0))); err != nil {
			return err
		}
	default:
		return paramBadType
	}

	return nil
}

func (v *FixedCOctetStringValue) Deserialize(buff *bytes.Buffer) error {
	octetStringValue := NewOctetStringValue(v.maxLen)
	if err := octetStringValue.Deserialize(buff); err != nil {
		return err
	} else if l := len(v.raw); l != 1 || l != v.maxLen {
		return fmt.Errorf("real length %v not equal 1 or %v fixed length", l, v.maxLen)
	} else {
		v.OctetStringValue = octetStringValue
		return nil
	}
}

type Uint8Value struct {
	*OctetStringValue
}

func NewUint8Value() *Uint8Value {
	return &Uint8Value{OctetStringValue: NewOctetStringValue(1)}
}

func (v *Uint8Value) String() string {
	return fmt.Sprintf("%v", v.raw[0])
}

func (v *Uint8Value) Set(d interface{}) error {
	switch data := d.(type) {
	case uint8:
		v.raw[0] = data
	case int:
		v.raw[0] = uint8(data)
	default:
		return paramBadType
	}

	return nil
}

func (v *Uint8Value) Deserialize(buff *bytes.Buffer) error {
	if b, err := buff.ReadByte(); err != nil {
		return err
	} else {
		v.raw[0] = b
		return nil
	}
}

func (v *Uint8Value) Uint32() (uint32, error) {
	return uint32(v.raw[0]), nil
}

type Uint16Value struct {
	*OctetStringValue
}

func NewUint16Value() *Uint16Value {
	return &Uint16Value{OctetStringValue: NewOctetStringValue(2)}
}

func (v *Uint16Value) String() string {
	return fmt.Sprintf("%v", binary.BigEndian.Uint16(v.raw))
}

func (v *Uint16Value) Set(d interface{}) error {
	switch data := d.(type) {
	case uint8:
		binary.BigEndian.PutUint16(v.raw, uint16(data))
	case uint16:
		binary.BigEndian.PutUint16(v.raw, data)
	case int:
		binary.BigEndian.PutUint16(v.raw, uint16(data))
	default:
		return paramBadType
	}

	return nil
}

func (v *Uint16Value) Deserialize(buff *bytes.Buffer) error {
	bytes := make([]byte, v.Len())
	if n, err := buff.Read(bytes); n != len(bytes) {
		return fmt.Errorf("readed %v bytes, need %v", n, len(bytes))
	} else if err != nil {
		return err
	} else {
		v.raw = bytes
		return nil
	}
}

func (v *Uint16Value) Uint32() (uint32, error) {
	return uint32(binary.BigEndian.Uint16(v.raw)), nil
}

type Uint32Value struct {
	*OctetStringValue
}

func NewUint32Value() *Uint16Value {
	return &Uint16Value{OctetStringValue: NewOctetStringValue(4)}
}

func (v *Uint32Value) String() string {
	return fmt.Sprintf("%v", binary.BigEndian.Uint32(v.raw))
}

func (v *Uint32Value) Set(d interface{}) error {
	switch data := d.(type) {
	case uint32:
		binary.BigEndian.PutUint32(v.raw, data)
	case uint8:
		binary.BigEndian.PutUint32(v.raw, uint32(data))
	case uint16:
		binary.BigEndian.PutUint32(v.raw, uint32(data))
	case int:
		binary.BigEndian.PutUint32(v.raw, uint32(data))
	default:
		return paramBadType
	}

	return nil
}

func (v *Uint32Value) Deserialize(buff *bytes.Buffer) error {
	bytes := make([]byte, v.Len())
	if n, err := buff.Read(bytes); n != len(bytes) {
		return fmt.Errorf("readed %v bytes, need %v", n, len(bytes))
	} else if err != nil {
		return err
	} else {
		v.raw = bytes
		return nil
	}
}

func (v *Uint32Value) Uint32() (uint32, error) {
	return binary.BigEndian.Uint32(v.raw), nil
}

func (t Tag) String() string {
	switch t {
	case TagDestAddrSubunit:
		return "dest_addr_subunit"
	case TagDestNetworkType:
		return "dest_network_type"
	case TagDestBearerType:
		return "dest_bearer_type"
	case TagDestTelematicsID:
		return "dest_telematics_id"
	case TagSourceAddrSubunit:
		return "source_addr_subunit"
	case TagSourceNetworkType:
		return "source_network_type"
	case TagSourceBearerType:
		return "source_bearer_type"
	case TagSourceTelematicsID:
		return "source_telematics_id"
	case TagQosTimeToLive:
		return "qos_time_to_live"
	case TagPayloadType:
		return "payload_type"
	case TagAdditionalStatusInfoText:
		return "additional_status_info_text"
	case TagReceiptedMessageID:
		return "receipted_message_id"
	case TagMsMsgWaitFacilities:
		return "ms_msg_wait_facilities"
	case TagPrivacyIndicator:
		return "privacy_indicator"
	case TagSourceSubaddress:
		return "source_subaddress"
	case TagDestSubaddress:
		return "dest_subaddress"
	case TagUserMessageReference:
		return "user_message_reference"
	case TagUserResponseCode:
		return "user_response_code"
	case TagSourcePort:
		return "source_port"
	case TagDestinationPort:
		return "destination_port"
	case TagSarMsgRefNum:
		return "sar_msg_ref_num"
	case TagLanguageIndicator:
		return "language_indicator"
	case TagSarTotalSegments:
		return "sar_total_segments"
	case TagSarSegmentSeqnum:
		return "sar_segment_seqnum"
	case TagCallbackNumPresInd:
		return "callback_num_pres_ind"
	case TagCallbackNumAtag:
		return "callback_num_atag"
	case TagNumberOfMessages:
		return "number_od_messages"
	case TagCallbackNum:
		return "callback_num"
	case TagDpfResult:
		return "dpf_result"
	case TagSetDpf:
		return "set_dpf"
	case TagMsAvailabilityStatus:
		return "ms_availability_status"
	case TagNetworkErrorCode:
		return "network_error_code"
	case TagMessagePayload:
		return "message_payload"
	case TagDeliveryFailureReason:
		return "delivery_failure_reason"
	case TagMoreMessagesToSend:
		return "more_messages_to_send"
	case TagMessageStateOption:
		return "message_state"
	case TagUssdServiceOp:
		return "ussd_service_op"
	case TagDisplayTime:
		return "display_time"
	case TagSmsSignal:
		return "sms_signal"
	case TagMsValidity:
		return "ms_validity"
	case TagAlertOnMessageDelivery:
		return "alert_on_message_delivery"
	case TagItsReplyType:
		return "its_reply_type"
	case TagItsSessionInfo:
		return "its_session_info"
	default:
		return fmt.Sprintf("unknown tlv (%d)", t)
	}
}
