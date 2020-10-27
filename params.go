package zkm

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

type mandatoryParams struct {
	names  []Name
	params map[Name]*mandatoryParam
}

func newMandatoryParams(id Id) *mandatoryParams {
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

	params := make(map[Name]*mandatoryParam)
	for _, n := range names {
		params[n] = newMandatoryParam(n)
	}

	return &mandatoryParams{names: names, params: params}
}

func (ps *mandatoryParams) len() uint32 {
	totalLen := 0
	for _, p := range ps.params {
		totalLen += p.value().len()
	}

	return uint32(totalLen)
}

func (ps *mandatoryParams) serialize() []byte {
	buff := bytes.Buffer{}
	for _, n := range ps.names {
		buff.Write(ps.params[n].value().raw())
	}
	return buff.Bytes()
}

func (ps *mandatoryParams) deserialize(buff *bytes.Buffer) error {
	for _, n := range ps.names {
		if err := ps.params[n].value().deserialize(buff); err != nil {
			return err
		}

		if n == SMLength {
			smLength, err := ps.params[n].value().uint32()
			if err != nil {
				panic(err)
			}
			ps.params[ShortMessage].v = newOctetStringValue(int(smLength))
		}
	}
	return nil
}

func (ps *mandatoryParams) get(name Name) (*mandatoryParam, error) {
	if p, ok := ps.params[name]; ok {
		return p, nil
	} else {
		return nil, paramNotFound
	}
}

type mandatoryParam struct {
	n Name
	v value
}

func newMandatoryParam(name Name) *mandatoryParam {
	var value value
	switch name {
	case InterfaceVersion, AddrTON, AddrNPI, SourceAddrTON, SourceAddrNPI, DestAddrTON, DestAddrNPI, EsmeAddrTON,
		EsmeAddrNPI, ESMClass, ProtocolID, PriorityFlag, RegisteredDelivery, ReplaceIfPresentFlag, DataCoding,
		SMDefaultMsgID, SMLength, NumberDests, DestFlag, NoUnsuccess, MessageState, ErrorCode:
		value = newUint8Value()
	case ScheduleDeliveryTime, ValidityPeriod, FinalDate:
		value = newFixedCOctetStringValue(17)
	case SourceAddr, DestinationAddr, DlName:
		value = newCOctetStringValue(21)
	case ShortMessage:
		value = newOctetStringValue(0)
	case SystemID:
		value = newCOctetStringValue(16)
	case Password:
		value = newCOctetStringValue(9)
	case SystemType:
		value = newCOctetStringValue(13)
	case AddressRange:
		value = newCOctetStringValue(41)
	case MessageID, EsmeAddr:
		value = newCOctetStringValue(65)
	case ServiceType:
		value = newCOctetStringValue(6)
	default:
		value = newOctetStringValue(0)
	}

	return &mandatoryParam{n: name, v: value}
}

func (p *mandatoryParam) String() string {
	return fmt.Sprintf("%v: %v", p.n, p.v)
}

func (p *mandatoryParam) name() Name {
	return p.n
}

func (p *mandatoryParam) value() value {
	return p.v
}

type optionalParam struct {
	t Tag
	v value
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	} else {
		return b
	}
}

type optionalParams struct {
	params map[Tag]*optionalParam
}

func newOptionalParams() *optionalParams {
	return &optionalParams{params: make(map[Tag]*optionalParam)}
}

func (ps *optionalParams) len() uint32 {
	totalLen := 0
	for _, p := range ps.params {
		totalLen += p.len()
	}

	return uint32(totalLen)
}

func (ps *optionalParams) serialize() []byte {
	buff := bytes.Buffer{}
	b := make([]byte, 2)
	for _, p := range ps.params {
		binary.BigEndian.PutUint16(b, uint16(p.t))
		buff.Write(b)
		binary.BigEndian.PutUint16(b, uint16(p.value().len()))
		buff.Write(b)
		buff.Write(p.value().raw())
	}
	return buff.Bytes()
}

func (ps *optionalParams) deserialize(buff *bytes.Buffer) error {
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
		optParam := newOptionalParam(tag, l)

		if err = optParam.value().deserialize(buff); err != nil {
			return err
		}

		if l != uint16(optParam.value().len()) {
			return fmt.Errorf("bad optional param: len %v, real len %v", l, optParam.value().len())
		}

		ps.params[tag] = optParam
	}

	return nil
}

func (ps *optionalParams) get(tag Tag) (*optionalParam, error) {
	if p, ok := ps.params[tag]; ok {
		return p, nil
	} else {
		return nil, paramNotFound
	}
}

func (ps *optionalParams) add(tag Tag, d interface{}) error {
	p := newOptionalParam(tag, 0)
	if err := p.value().set(d); err != nil {
		return err
	}
	ps.params[tag] = p
	return nil
}

func (ps *optionalParams) remove(tag Tag) {
	delete(ps.params, tag)
}

func newOptionalParam(tag Tag, len uint16) *optionalParam {
	var value value
	switch tag {
	case TagDestAddrSubunit, TagDestNetworkType, TagSourceNetworkType, TagDestBearerType, TagSourceBearerType,
		TagSourceAddrSubunit, TagSourceTelematicsID, TagPayloadType, TagMsMsgWaitFacilities, TagPrivacyIndicator,
		TagUserResponseCode, TagLanguageIndicator, TagSarTotalSegments, TagSarSegmentSeqnum, TagScInterfaceVersion,
		TagDisplayTime, TagMsValidity, TagDpfResult, TagSetDpf, TagMsAvailabilityStatus, TagDeliveryFailureReason,
		TagMoreMessagesToSend, TagMessageStateOption, TagCallbackNumPresInd, TagNumberOfMessages, TagItsReplyType, TagUssdServiceOp:
		value = newUint8Value()
	case TagDestTelematicsID, TagUserMessageReference, TagSourcePort, TagDestinationPort,
		TagSarMsgRefNum, TagSmsSignal, TagItsSessionInfo:
		value = newUint16Value()
	case TagQosTimeToLive:
		value = newUint32Value()
	case TagAdditionalStatusInfoText:
		value = newCOctetStringValue(256)
	case TagReceiptedMessageID:
		value = newCOctetStringValue(65)
	case TagSourceSubaddress, TagDestSubaddress:
		value = newOctetStringValue(int(min(23, len)))
	case TagNetworkErrorCode:
		value = newOctetStringValue(3)
	case TagMessagePayload:
		value = newOctetStringValue(int(min(math.MaxUint16, len)))
	case TagCallbackNum:
		value = newOctetStringValue(int(min(19, len)))
	default:
		value = newOctetStringValue(0)
	}

	return &optionalParam{t: tag, v: value}
}

func (p *optionalParam) String() string {
	return fmt.Sprintf("%v: %v", p.t, p.v)
}

func (p *optionalParam) tag() Tag {
	return p.t
}

func (p *optionalParam) len() int {
	return 4 + p.v.len()
}

func (p *optionalParam) value() value {
	return p.v
}

type value interface {
	fmt.Stringer
	len() int
	set(d interface{}) error
	deserialize(buff *bytes.Buffer) error
	raw() []byte
	uint32() (uint32, error)
}

type octetStringValue struct {
	r []byte
}

func newOctetStringValue(len int) *octetStringValue {
	return &octetStringValue{r: make([]byte, len)}
}

func (v *octetStringValue) String() string {
	return fmt.Sprintf("%v", v.r)
}

func (v *octetStringValue) len() int {
	return len(v.r)
}

func (v *octetStringValue) set(d interface{}) error {
	switch data := d.(type) {
	case []byte:
		v.r = data
	default:
		return paramBadType
	}

	return nil
}

func (v *octetStringValue) deserialize(buff *bytes.Buffer) error {
	b := make([]byte, v.len())
	if n, _ := buff.Read(b); n == v.len() {
		v.r = b
		return nil
	} else {
		return fmt.Errorf("readed %v bytes, need %v", n, len(v.r))
	}
}

func (v *octetStringValue) raw() []byte {
	return v.r
}

func (v *octetStringValue) uint32() (uint32, error) {
	return 0, paramBadType
}

type cOctetStringValue struct {
	*octetStringValue
	maxLen int
}

func (v *cOctetStringValue) String() string {
	return string(v.r[:len(v.r)-1])
}

func (v *cOctetStringValue) set(d interface{}) error {
	switch data := d.(type) {
	case string:
		if err := v.deserialize(bytes.NewBuffer(append([]byte(data), 0))); err != nil {
			return err
		}
	default:
		return paramBadType
	}

	return nil
}

func (v *cOctetStringValue) deserialize(buff *bytes.Buffer) error {
	if b, err := buff.ReadBytes(0x00); err != nil {
		return err
	} else if len(b) > v.maxLen {
		return fmt.Errorf("real length %v exceeded the maximum length %v", len(v.r), v.maxLen)
	} else {
		v.r = b
		return nil
	}
}

func newCOctetStringValue(maxLen int) *cOctetStringValue {
	return &cOctetStringValue{octetStringValue: newOctetStringValue(1), maxLen: maxLen}
}

type fixedCOctetStringValue struct {
	*cOctetStringValue
}

func newFixedCOctetStringValue(len int) *fixedCOctetStringValue {
	return &fixedCOctetStringValue{cOctetStringValue: newCOctetStringValue(len)}
}

func (v *fixedCOctetStringValue) set(d interface{}) error {
	switch data := d.(type) {
	case string:
		if err := v.deserialize(bytes.NewBuffer(append([]byte(data), 0))); err != nil {
			return err
		}
	default:
		return paramBadType
	}

	return nil
}

func (v *fixedCOctetStringValue) deserialize(buff *bytes.Buffer) error {
	octetStringValue := newCOctetStringValue(v.maxLen)
	if err := octetStringValue.deserialize(buff); err != nil {
		return err
	} else if l := len(octetStringValue.r); l != 1 && l != v.maxLen {
		return fmt.Errorf("real length %v not equal 1 or %v fixed length", l, v.maxLen)
	} else {
		v.cOctetStringValue = octetStringValue
		return nil
	}
}

type uint8Value struct {
	*octetStringValue
}

func newUint8Value() *uint8Value {
	return &uint8Value{octetStringValue: newOctetStringValue(1)}
}

func (v *uint8Value) String() string {
	return fmt.Sprintf("%v", v.r[0])
}

func (v *uint8Value) set(d interface{}) error {
	switch data := d.(type) {
	case uint8:
		v.r[0] = data
	case int:
		v.r[0] = uint8(data)
	default:
		return paramBadType
	}

	return nil
}

func (v *uint8Value) deserialize(buff *bytes.Buffer) error {
	if b, err := buff.ReadByte(); err != nil {
		return err
	} else {
		v.r[0] = b
		return nil
	}
}

func (v *uint8Value) uint32() (uint32, error) {
	return uint32(v.r[0]), nil
}

type uint16Value struct {
	*octetStringValue
}

func newUint16Value() *uint16Value {
	return &uint16Value{octetStringValue: newOctetStringValue(2)}
}

func (v *uint16Value) String() string {
	return fmt.Sprintf("%v", binary.BigEndian.Uint16(v.r))
}

func (v *uint16Value) set(d interface{}) error {
	switch data := d.(type) {
	case uint8:
		binary.BigEndian.PutUint16(v.r, uint16(data))
	case uint16:
		binary.BigEndian.PutUint16(v.r, data)
	case int:
		binary.BigEndian.PutUint16(v.r, uint16(data))
	default:
		return paramBadType
	}

	return nil
}

func (v *uint16Value) deserialize(buff *bytes.Buffer) error {
	b := make([]byte, v.len())
	if n, err := buff.Read(b); n != len(b) {
		return fmt.Errorf("readed %v bytes, need %v", n, len(b))
	} else if err != nil {
		return err
	} else {
		v.r = b
		return nil
	}
}

func (v *uint16Value) uint32() (uint32, error) {
	return uint32(binary.BigEndian.Uint16(v.r)), nil
}

type uint32Value struct {
	*octetStringValue
}

func newUint32Value() *uint32Value {
	return &uint32Value{octetStringValue: newOctetStringValue(4)}
}

func (v *uint32Value) String() string {
	return fmt.Sprintf("%v", binary.BigEndian.Uint32(v.r))
}

func (v *uint32Value) set(d interface{}) error {
	switch data := d.(type) {
	case uint32:
		binary.BigEndian.PutUint32(v.r, data)
	case uint8:
		binary.BigEndian.PutUint32(v.r, uint32(data))
	case uint16:
		binary.BigEndian.PutUint32(v.r, uint32(data))
	case int:
		binary.BigEndian.PutUint32(v.r, uint32(data))
	default:
		return paramBadType
	}

	return nil
}

func (v *uint32Value) deserialize(buff *bytes.Buffer) error {
	b := make([]byte, v.len())
	if n, err := buff.Read(b); n != len(b) {
		return fmt.Errorf("readed %v bytes, need %v", n, len(b))
	} else if err != nil {
		return err
	} else {
		v.r = b
		return nil
	}
}

func (v *uint32Value) uint32() (uint32, error) {
	return binary.BigEndian.Uint32(v.r), nil
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
	case TagScInterfaceVersion:
		return "sc_interface_version"
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
