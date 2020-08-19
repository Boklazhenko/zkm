package mandatory

type Name string

const (
	AddrNPI              Name = "addr_npi"
	AddrTON              Name = "addr_ton"
	AddressRange         Name = "address_range"
	DataCoding           Name = "data_coding"
	DestAddrNPI          Name = "dest_addr_npi"
	DestAddrTON          Name = "dest_addr_ton"
	DestinationAddr      Name = "destination_addr"
	DestinationList      Name = "dest_addresses"
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
	UDHLength            Name = "gsm_sms_ud.udh.len"
	GSMUserData          Name = "gsm_sms_ud.udh"
	UnsuccessSme         Name = "unsuccess_sme"
	ValidityPeriod       Name = "validity_period"
)

type Params map[Name]Value

type Value interface {
}
