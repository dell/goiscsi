package goiscsi

// ISCSITarget defines an iSCSI target
type ISCSITarget struct {
	Portal   string
	GroupTag string
	Target   string
}

type ISCSISessionState string
type ISCSIConnectionState string
type ISCSITransportName string

const (
	ISCSISessionState_LOGGED_IN ISCSISessionState = "LOGGED_IN"
	ISCSISessionState_FAILED    ISCSISessionState = "FAILED"
	ISCSISessionState_FREE      ISCSISessionState = "FREE"

	ISCSIConnectionState_FREE             ISCSIConnectionState = "FREE"
	ISCSIConnectionState_TRANSPORT_WAIT   ISCSIConnectionState = "TRANSPORT WAIT"
	ISCSIConnectionState_IN_LOGIN         ISCSIConnectionState = "IN LOGIN"
	ISCSIConnectionState_LOGGED_IN        ISCSIConnectionState = "LOGGED IN"
	ISCSIConnectionState_IN_LOGOUT        ISCSIConnectionState = "IN LOGOUT"
	ISCSIConnectionState_LOGOUT_REQUESTED ISCSIConnectionState = "LOGOUT REQUESTED"
	ISCSIConnectionState_CLEANUP_WAIT     ISCSIConnectionState = "CLEANUP WAIT"

	ISCSITransportName_TCP  ISCSITransportName = "tcp"
	ISCSITransportName_ISER ISCSITransportName = "iser"
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

// ISCSISession defines an iSCSI node info
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
