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
