package optional

import "fmt"

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

type Params map[Tag]Value

type Value interface {
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
