package goiscsi

import (
	"errors"
)

// ISCSIinterface is the interface that provides the iSCSI client functionality
type ISCSIinterface interface {
	// Discover the targets exposed via a given portal
	// returns an array of ISCSITarget instances
	DiscoverTargets(address string, login bool) ([]ISCSITarget, error)

	// Get a list of iSCSI initiators defined in a specified file
	// To use the system default file of "/etc/iscsi/initiatorname.iscsi", provide a filename of ""
	GetInitiators(filename string) ([]string, error)

	// Log into a specified target
	PerformLogin(target ISCSITarget) error

	// Log out of a specified target
	PerformLogout(target ISCSITarget) error

	// Rescan current iSCSI sessions
	PerformRescan() error

	// Query information about sessions
	GetSessions() ([]ISCSISession, error)

	// Query information about nodes
	GetNodes() ([]ISCSINode, error)

	// Set CHAP credentials for a target (creates/updates node database)
	SetCHAPCredentials(target ISCSITarget, username, password string) error

	// CreateOrUpdateNode creates new or update existing iSCSI node in iscsid database
	CreateOrUpdateNode(target ISCSITarget, options map[string]string) error

	// DeleteNode delete iSCSI node from iscsid database
	DeleteNode(target ISCSITarget) error

	// generic implementations
	isMock() bool
	getOptions() map[string]string
}

// ISCSIType is the base structre for each platform implementation
type ISCSIType struct {
	mock    bool
	options map[string]string
}

var (
	// ErrIscsiNotInstalled is returned when the iscsi utilities are not
	// found on a system
	ErrIscsiNotInstalled = errors.New("iSCSI utilities are not installed")
	// ErrNotImplemented is returned when a platform does not implement
	ErrNotImplemented = errors.New("not implemented")
)

func (i *ISCSIType) isMock() bool {
	return i.mock
}

func (i *ISCSIType) getOptions() map[string]string {
	return i.options
}
