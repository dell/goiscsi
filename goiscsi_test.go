// +build linux darwin

package goiscsi

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	testPortal string
	testTarget string
)

func reset() {
	if p := os.Getenv("GOISCSI_PORTAL"); p != "" {
		testPortal = p
	}
	if t := os.Getenv("GOISCSI_TARGET"); t != "" {
		testTarget = t
	}
	GOISCSIMock.InduceDiscoveryError = false
	GOISCSIMock.InduceInitiatorError = false
	GOISCSIMock.InduceLoginError = false
	GOISCSIMock.InduceLogoutError = false
	GOISCSIMock.InduceRescanError = false
}

func TestPolymorphichCapability(t *testing.T) {
	reset()
	var c ISCSIinterface
	// start off with a real implementation
	c = NewLinuxISCSI(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
	// switch it to mock
	c = NewMockISCSI(map[string]string{})
	if !c.isMock() {
		// this should not be a real implementation
		t.Error("Expected a mock implementation but got a real one")
		return
	}
	// switch back to a real implementation
	c = NewLinuxISCSI(map[string]string{})
	if c.isMock() {
		// this should not be a mock implementation
		t.Error("Expected a real implementation but got a mock one")
		return
	}
}

func TestDiscoverTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	_, err := c.DiscoverTargets(testPortal, false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	err := c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformLogout(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestLoginLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	err := c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformLogout(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestLogoutLogoutTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	// log out of the target, just in case we are logged in already
	_ = c.PerformLogin(tgt)
	err := c.PerformLogout(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}
func TestGetInitiators(t *testing.T) {
	reset()
	var testdata = []struct {
		filename string
		count    int
	}{
		{"testdata/initiatorname.iscsi", 1},
		{"testdata/multiple_iqn.iscsi", 2},
		{"testdata/no_iqn.iscsi", 0},
		{"testdata/valid.iscsi", 1},
		{"testdata/with_comments.iscsi", 1},
	}

	c := NewLinuxISCSI(map[string]string{})
	for _, tt := range testdata {
		initiators, err := c.GetInitiators(tt.filename)
		if err != nil {
			t.Errorf("Error getting %d initiators from %s: %s", tt.count, tt.filename, err.Error())
		}
		if len(initiators) != tt.count {
			t.Errorf("Expected %d initiators in %s, but got %d", tt.count, tt.filename, len(initiators))
		}
	}

}

func TestPerformRescan(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	err := c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformRescan()
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestBuildISCSICommand(t *testing.T) {
	reset()
	opts := map[string]string{}
	initial := []string{"/bin/ls"}
	opts[ChrootDirectory] = "/test"
	c := NewLinuxISCSI(opts)
	command := c.buildISCSICommand(initial)
	// the length of the resulting command should the length of the initial command +2
	if len(command) != (len(initial) + 2) {
		t.Errorf("Expected to %d items in the command slice but received %v", len(initial)+2, command)
	}
	if command[0] != "chroot" {
		t.Error("Expected the command to be run with chroot")
	}
	if command[1] != opts[ChrootDirectory] {
		t.Errorf("Expected the command to chroot to %s but got %s", opts[ChrootDirectory], command[1])
	}
}

func TestMockDiscoverTargets(t *testing.T) {
	reset()
	var c ISCSIinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTargets] = fmt.Sprintf("%d", expected)
	c = NewMockISCSI(opts)
	//c = mock
	targets, err := c.DiscoverTargets("10.0.1.2", true)
	if err != nil {
		t.Error(err.Error())
	}
	if len(targets) != expected {
		t.Errorf("Expected to find %d targets, but got back %v", expected, targets)
	}
}

func TestMockDiscoverTargetsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTargets] = fmt.Sprintf("%d", expected)
	c := NewMockISCSI(opts)
	GOISCSIMock.InduceDiscoveryError = true
	targets, err := c.DiscoverTargets("10.0.1.2", false)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(targets) != 0 {
		t.Errorf("Expected to receive 0 targets when inducing an error. Received %v", targets)
		return
	}
}

func TestMockGetInitiators(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockISCSI(opts)
	initiators, err := c.GetInitiators("")
	if err != nil {
		t.Error(err.Error())
	}
	if len(initiators) != expected {
		t.Errorf("Expected to find %d initiators, but got back %v", expected, initiators)
	}
}

func TestMockGetInitiatorsError(t *testing.T) {
	reset()
	opts := map[string]string{}
	expected := 3
	opts[MockNumberOfInitiators] = fmt.Sprintf("%d", expected)
	c := NewMockISCSI(opts)
	GOISCSIMock.InduceInitiatorError = true
	initiators, err := c.GetInitiators("")
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
	if len(initiators) != 0 {
		t.Errorf("Expected to receive 0 initiators when inducing an error. Received %v", initiators)
		return
	}
}

func TestMockLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	err := c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformLogout(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockLogoutTargetsError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	GOISCSIMock.InduceLogoutError = true
	err := c.PerformLogin(tgt)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = c.PerformLogout(tgt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockLoginTargetsError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   testPortal,
		GroupTag: "0",
		Target:   testTarget,
	}
	GOISCSIMock.InduceLoginError = true
	err := c.PerformLogin(tgt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockPerformRescan(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check without induced error
	err := c.PerformRescan()
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockPerformRescanError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceRescanError = true
	err := c.PerformRescan()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}
