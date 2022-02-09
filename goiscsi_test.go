//go:build linux || darwin
// +build linux darwin

package goiscsi

import (
	"fmt"
	"io/ioutil"
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
	GOISCSIMock.InduceGetSessionsError = false
	GOISCSIMock.InduceGetNodesError = false
	GOISCSIMock.InduceCreateOrUpdateNodeError = false
	GOISCSIMock.InduceSetCHAPError = false
	GOISCSIMock.InduceDeleteNodeError = false
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

func TestGetSessions(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	_, err := c.GetSessions()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNodes(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	_, err := c.GetNodes()
	if err != nil {
		t.Error(err.Error())
	}
}

func TestCreateOrUpdateNode(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal: "foo",
		Target: "bar"}
	opt := make(map[string]string)
	err := c.CreateOrUpdateNode(tgt, opt)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestDeleteNode(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal: "foo",
		Target: "bar"}
	err := c.DeleteNode(tgt)
	if err != nil {
		t.Error(err.Error())
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
	targets, err := c.DiscoverTargets("1.1.1.1", true)
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
	targets, err := c.DiscoverTargets("1.1.1.1", false)
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

func TestMockGetSessions(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check without induced error
	data, err := c.GetSessions()
	if len(data) == 0 || len(data[0].Target) == 0 {
		t.Error("invalid response from mock")
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockGetSessionsError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceGetSessionsError = true
	_, err := c.GetSessions()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockGetNodes(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check without induced error
	data, err := c.GetNodes()
	if len(data) == 0 || len(data[0].Target) == 0 {
		t.Error("invalid response from mock")
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockGetNodesError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceGetNodesError = true
	_, err := c.GetNodes()
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockCreateOrUpdateNode(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check without induced error
	opt := make(map[string]string)
	err := c.CreateOrUpdateNode(ISCSITarget{}, opt)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockCreateOrUpdateNodeError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceCreateOrUpdateNodeError = true
	opt := make(map[string]string)
	err := c.CreateOrUpdateNode(ISCSITarget{}, opt)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockSetCHAPCredentials(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceSetCHAPError = true
	username := "username"
	chapSecret := "secret"
	err := c.SetCHAPCredentials(ISCSITarget{}, username, chapSecret)
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestMockDeleteNode(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check without induced error
	err := c.DeleteNode(ISCSITarget{})
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestMockDeleteNodeError(t *testing.T) {
	reset()
	c := NewMockISCSI(map[string]string{})
	// check with induced error
	GOISCSIMock.InduceDeleteNodeError = true
	err := c.DeleteNode(ISCSITarget{})
	if err == nil {
		t.Error("Expected an induced error")
		return
	}
	if !strings.Contains(err.Error(), "induced") {
		t.Error("Expected an induced error")
		return
	}
}

func TestSessionParserParse(t *testing.T) {
	sp := &sessionParser{}
	fileErrMsg := "can't read file with test data"

	// test valid data
	data, err := ioutil.ReadFile("testdata/session_info_valid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	sessions := sp.Parse(data)
	if len(sessions) != 2 {
		t.Error("unexpected results count")
	}
	for i, session := range sessions {
		if i == 0 {
			compareStr(t, session.Target, "iqn.2015-10.com.dell:dellemc-foobar-123-a-7ceb34a3")
			compareStr(t, session.Portal, "192.168.1.1:3260")
			compareStr(t, session.SID, "12")
			compareStr(t, string(session.IfaceTransport), string(ISCSITransportName_TCP))
			compareStr(t, session.IfaceInitiatorname, "iqn.1994-05.com.redhat:650e84b584d")
			compareStr(t, session.IfaceIPaddress, "1.1.1.1")
			compareStr(t, string(session.ISCSISessionState), string(ISCSISessionState_LOGGED_IN))
			compareStr(t, string(session.ISCSIConnectionState), string(ISCSIConnectionState_LOGGED_IN))
			compareStr(t, session.Username, "admin")
			compareStr(t, session.Password, "foobar")
			compareStr(t, session.UsernameIn, "")
			compareStr(t, session.PasswordIn, "")
		} else {
			compareStr(t, session.Target, "iqn.2015-10.com.dell:dellemc-foobar-123-b-61ecc53a")
			compareStr(t, session.Portal, "192.168.1.2:3260")
			compareStr(t, session.SID, "13")
			compareStr(t, string(session.IfaceTransport), string(ISCSITransportName_TCP))
			compareStr(t, session.IfaceInitiatorname, "iqn.1994-05.com.redhat:650e84b585d")
			compareStr(t, session.IfaceIPaddress, "1.1.1.1")
			compareStr(t, string(session.ISCSISessionState), string(ISCSISessionState_FAILED))
			compareStr(t, string(session.ISCSIConnectionState), string(ISCSIConnectionState_FREE))
			compareStr(t, session.Username, "")
			compareStr(t, session.Password, "")
			compareStr(t, session.UsernameIn, "")
			compareStr(t, session.PasswordIn, "")
		}
	}

	// test invalid data parsing
	data, err = ioutil.ReadFile("testdata/session_info_invalid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	r := sp.Parse(data)
	if len(r) != 0 {
		t.Error("non empty result while parsing invalid data")
	}
}

func TestNodeParserParse(t *testing.T) {
	np := &nodeParser{}
	fileErrMsg := "can't read file with test data"

	// test valid data
	data, err := ioutil.ReadFile("testdata/node_info_valid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	nodes := np.Parse(data)
	if len(nodes) != 2 {
		t.Error("unexpected results count")
	}
	for i, node := range nodes {
		if i == 0 {
			trgt := "iqn.2015-10.com.dell:dellemc-foobar-123-b-61ecc53a"
			compareStr(t, node.Target, trgt)
			compareStr(t, node.Portal, "192.168.1.2:3260")
			compareStr(t, node.Fields["node.name"], trgt)
			compareStr(t, node.Fields["node.conn[0].iscsi.OFMarker"], "No")
		} else {
			compareStr(t, node.Target, "iqn.2015-10.com.dell:dellemc-foobar-123-a-7ceb34a3")
			compareStr(t, node.Portal, "192.168.1.2:3260")
		}
	}

	// test invalid data parsing
	data, err = ioutil.ReadFile("testdata/node_info_invalid")
	if err != nil {
		t.Error(fileErrMsg)
	}
	r := np.Parse(data)
	if len(r) != 0 {
		t.Error("non empty result while parsing invalid data")
	}
}

func compareStr(t *testing.T, str1 string, str2 string) {
	if str1 != str2 {
		t.Errorf("strings are not equal: %s != %s", str1, str2)
	}
}

func TestFieldKeyValue(t *testing.T) {
	str1 := "node.name = iqn.2015-10.com.dell:dellemc-foobar-123-a-7ceb34a3"
	key, value := fieldKeyValue(str1, "=")
	if key != "node.name" {
		t.Error("invalid key")
	}
	if value != "iqn.2015-10.com.dell:dellemc-foobar-123-a-7ceb34a3" {
		t.Error("invalid value")
	}

	str2 := "iSCSI Connection State: LOGGED IN"
	key, value = fieldKeyValue(str2, ":")
	if key != "iSCSI Connection State" {
		t.Error("invalid key")
	}
	if value != "LOGGED IN" {
		t.Error("invalid value")
	}
}
