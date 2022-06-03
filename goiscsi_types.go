package goiscsi

// ISCSITarget defines an iSCSI target
type ISCSITarget struct {
	Portal   string
	GroupTag string
	Target   string
}

// ISCSISessionState holds iscsi session state
type ISCSISessionState string

// ISCSIConnectionState holds iscsi connection state
type ISCSIConnectionState string

// ISCSITransportName holds iscsi transport name
type ISCSITransportName string

// ISCSI session and connection states
const (
	// ISCSISessionState_LOGGED_IN has been deprecated and will be removed in future release
	ISCSISessionState_LOGGED_IN ISCSISessionState = "LOGGED_IN"
	// ISCSISessionState_FAILED has been deprecated and will be removed in future release
	ISCSISessionState_FAILED ISCSISessionState = "FAILED"
	// ISCSISessionState_FREE has been deprecated and will be removed in future release
	ISCSISessionState_FREE ISCSISessionState = "FREE"

	// ISCSIConnectionState_FREE has been deprecated and will be removed in future release
	ISCSIConnectionState_FREE ISCSIConnectionState = "FREE"
	// ISCSIConnectionState_TRANSPORT_WAIT has been deprecated and will be removed in future release
	ISCSIConnectionState_TRANSPORT_WAIT ISCSIConnectionState = "TRANSPORT WAIT"
	// ISCSIConnectionState_IN_LOGIN has been deprecated and will be removed in future release
	ISCSIConnectionState_IN_LOGIN ISCSIConnectionState = "IN LOGIN"
	// ISCSIConnectionState_LOGGED_IN has been deprecated and will be removed in future release
	ISCSIConnectionState_LOGGED_IN ISCSIConnectionState = "LOGGED IN"
	// ISCSIConnectionState_IN_LOGOUT has been deprecated and will be removed in future release
	ISCSIConnectionState_IN_LOGOUT ISCSIConnectionState = "IN LOGOUT"
	// ISCSIConnectionState_LOGOUT_REQUESTED has been deprecated and will be removed in future release
	ISCSIConnectionState_LOGOUT_REQUESTED ISCSIConnectionState = "LOGOUT REQUESTED"
	// ISCSIConnectionState_CLEANUP_WAIT has been deprecated and will be removed in future release
	ISCSIConnectionState_CLEANUP_WAIT ISCSIConnectionState = "CLEANUP WAIT"

	// ISCSITransportName_TCP has been deprecated and will be removed in future release
	ISCSITransportName_TCP ISCSITransportName = "tcp"
	// ISCSITransportName_ISER has been deprecated and will be removed in future release
	ISCSITransportName_ISER ISCSITransportName = "iser"

	ISCSISessionStateLOGGEDIN ISCSISessionState = "LOGGED_IN"
	ISCSISessionStateFAILED   ISCSISessionState = "FAILED"
	ISCSISessionStateFREE     ISCSISessionState = "FREE"

	ISCSIConnectionStateFREE            ISCSIConnectionState = "FREE"
	ISCSIConnectionStateTRANSPORTWAIT   ISCSIConnectionState = "TRANSPORT WAIT"
	ISCSIConnectionStateINLOGIN         ISCSIConnectionState = "IN LOGIN"
	ISCSIConnectionStateLOGGEDIN        ISCSIConnectionState = "LOGGED IN"
	ISCSIConnectionStateINLOGOUT        ISCSIConnectionState = "IN LOGOUT"
	ISCSIConnectionStateLOGOUTREQUESTED ISCSIConnectionState = "LOGOUT REQUESTED"
	ISCSIConnectionStateCLEANUPWAIT     ISCSIConnectionState = "CLEANUP WAIT"

	ISCSITransportNameTCP  ISCSITransportName = "tcp"
	ISCSITransportNameISER ISCSITransportName = "iser"
)

// ISCSISession defines an iSCSI session info
type ISCSISession struct {
	Target               string
	Portal               string
	SID                  string
	IfaceTransport       ISCSITransportName
	IfaceInitiatorname   string
	IfaceIPaddress       string
	ISCSISessionState    ISCSISessionState
	ISCSIConnectionState ISCSIConnectionState
	Username             string
	Password             string
	UsernameIn           string
	PasswordIn           string
}

// ISCSINode defines an iSCSI node info
type ISCSINode struct {
	Target string
	Portal string
	Fields map[string]string
}

type iSCSISessionParser interface {
	Parse([]byte) []ISCSISession
}

type iSCSINodeParser interface {
	Parse([]byte) []ISCSINode
}
