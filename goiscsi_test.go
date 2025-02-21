//go:build linux || darwin
// +build linux darwin

/*
 *
 * Copyright © 2019-2022 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package goiscsi

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	testPortal = "1.2.3.4"
	testTarget = "iqn.1991-05.com.emc:dummyExample"
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
	tests := []struct {
		name    string
		address string
		iface   string
		login   bool
		cmdOut  []byte
		cmdErr  error
		want    []ISCSITarget
		wantErr bool
	}{
		{
			name:    "valid address with target",
			address: "1.1.1.1",
			iface:   "iface0",
			login:   false,
			cmdOut:  []byte("1.1.1.1:3260,0 iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"),
			cmdErr:  nil,
			want: []ISCSITarget{
				{Portal: "1.1.1.1:3260", GroupTag: "0", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			},
			wantErr: false,
		},
		{
			name:    "error invalid address",
			address: "invalid address",
			iface:   "iface0",
			login:   false,
			cmdOut:  nil,
			cmdErr:  fmt.Errorf("invalid address"),
			want:    []ISCSITarget{},
			wantErr: true,
		},
		{
			name:    "error discovering targets",
			address: "1.1.1.1",
			iface:   "iface0",
			login:   false,
			cmdOut:  nil,
			cmdErr:  fmt.Errorf("iscsiadm error"),
			want:    []ISCSITarget{},
			wantErr: true,
		},
		{
			name:    "valid address without target",
			address: "2.2.2.2",
			iface:   "",
			login:   false,
			cmdOut:  []byte(""),
			cmdErr:  nil,
			want:    []ISCSITarget{},
			wantErr: false,
		},
		{
			name:    "multiple targets",
			address: "3.3.3.3",
			iface:   "iface0",
			login:   false,
			cmdOut:  []byte("3.3.3.3:3260,1 iqn.1992-04.com.emc:600009700bcbb70e3287017400000002\n3.3.3.3:3260,2 iqn.1992-04.com.emc:600009700bcbb70e3287017400000003"),
			cmdErr:  nil,
			want: []ISCSITarget{
				{Portal: "3.3.3.3:3260", GroupTag: "1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000002"},
				{Portal: "3.3.3.3:3260", GroupTag: "2", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000003"},
			},
			wantErr: false,
		},
		{
			name:    "interface specified without login",
			address: "4.4.4.4",
			iface:   "iface1",
			login:   false,
			cmdOut:  []byte("4.4.4.4:3260,0 iqn.1992-04.com.emc:600009700bcbb70e3287017400000004"),
			cmdErr:  nil,
			want: []ISCSITarget{
				{Portal: "4.4.4.4:3260", GroupTag: "0", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000004"},
			},
			wantErr: false,
		},
		{
			name:    "error on interface specified",
			address: "invalid",
			iface:   "iface0",
			login:   false,
			cmdOut:  nil,
			cmdErr:  fmt.Errorf("iscsiadm error"),
			want:    []ISCSITarget{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of LinuxISCSI
			reset()
			iscsi := NewLinuxISCSI(map[string]string{})

			// Set up mocks or other dependencies
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return tt.cmdOut, tt.cmdErr
			}

			// Call the discoverTargets function
			got, err := iscsi.discoverTargets(tt.address, tt.iface, tt.login)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("discoverTargets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if the result matches the expected result
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("discoverTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverUnreachableTargets(t *testing.T) {
	c := NewLinuxISCSI(map[string]string{})
	timeBeforeTestStart := time.Now()
	_, err := c.DiscoverTargets("127.0.0.1", false)
	timeAftertest := time.Now()
	// response should come within Timeout + 2 seconds
	if err != nil && (timeAftertest.Sub(timeBeforeTestStart).Seconds() > Timeout+2) {
		t.Error(err.Error())
	}
}

func TestDiscoverTargetsWithInterface(t *testing.T) {
	reset()
	iscsi := NewLinuxISCSI(map[string]string{})

	tests := []struct {
		name        string
		address     string
		iface       string
		login       bool
		cmdOut      []byte
		cmdErr      error
		expectedErr string
	}{
		{
			name:        "test empty address",
			address:     "",
			iface:       "iface0",
			login:       false,
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "error invalid IP or portal address",
		},
		{
			name:        "test iscsiadm not in path",
			address:     "1.1.1.1",
			iface:       "iface0",
			login:       false,
			cmdOut:      nil,
			cmdErr:      fmt.Errorf("exec: \"iscsiadm\": executable file not found in $PATH"),
			expectedErr: "exec: \"iscsiadm\": executable file not found in $PATH",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return test.cmdOut, test.cmdErr
			}
			targets, err := iscsi.DiscoverTargetsWithInterface(test.address, test.iface, test.login)
			if err != nil {
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			} else {
				// if test doesn't expect an error but returns targets length is 0, fail.
				if len(targets) == 0 && test.expectedErr != "" {
					t.Errorf("Expected error: %v, but got no error", test.expectedErr)
				}
			}
		})
	}
}

func TestLoginLogoutTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   "1.1.1.1", // Adjust this based on your `testPortal` variable
		GroupTag: "0",
		Target:   "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001", // Adjust this based on your `testTarget` variable
	}

	// Mock the runCommand function to simulate environment where iscsiadm is not found
	runCommand = func(_ *exec.Cmd) ([]byte, error) {
		return nil, errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	}

	expectedError := "exec: \"iscsiadm\": executable file not found in $PATH"

	// Test PerformLogin
	err := c.PerformLogin(tgt)
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
	}

	// Test PerformLogout
	err = c.PerformLogout(tgt)
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
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
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
		return
	}
	err = c.PerformLogin(tgt)
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
		return
	}
	err = c.PerformLogout(tgt)
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
		return
	}
}

func TestLoginUnreachableTargets(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal:   "127.0.0.1",
		GroupTag: "0",
		Target:   "iqn.1991-05.com.emc:dummyExample",
	}
	timeBeforeTestStart := time.Now()
	err := c.PerformLogin(tgt)
	timeAftertest := time.Now()
	// response should come within Timeout + 2 seconds
	if err != nil && (timeAftertest.Sub(timeBeforeTestStart).Seconds() > Timeout+2) {
		t.Error(err.Error())
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
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
		return
	}
}

func TestGetInitiators(t *testing.T) {
	reset()
	testdata := []struct {
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
	_, err := c.GetInitiators("")
	expectedError := errors.New("stat /etc/iscsi/initiatorname.iscsi: no such file or directory")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
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
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
		return
	}
	err = c.PerformRescan()
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
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
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
	}
}

func TestGetNodes(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	_, err := c.GetNodes()
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
	}
}

func TestCreateOrUpdateNode(t *testing.T) {
	iscsi := &LinuxISCSI{}

	tests := []struct {
		name        string
		target      ISCSITarget
		options     map[string]string
		cmdOut      []byte
		cmdErr      error
		expectedErr string
	}{
		{
			name:        "test invalid portal address",
			target:      ISCSITarget{Portal: "invalid", Target: "valid.iqn"},
			options:     map[string]string{},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "error invalid IP or portal address",
		},
		{
			name:        "test invalid IQN target",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "invalid"},
			options:     map[string]string{},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "error invalid IQN",
		},
		{
			name:        "test iscsiadm node not found",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			options:     map[string]string{},
			cmdOut:      nil,
			cmdErr:      simulateExitCode(21), // simulate "no objects found" exit code
			expectedErr: `exec: "iscsiadm": executable file not found in $PATH`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return test.cmdOut, test.cmdErr
			}
			err := iscsi.CreateOrUpdateNode(test.target, test.options)
			if err != nil {
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			} else {
				if test.expectedErr != "" {
					t.Errorf("Expected error: %v, but got no error", test.expectedErr)
				}
			}
		})
	}
}

func TestDeleteNode(t *testing.T) {
	iscsi := &LinuxISCSI{}

	tests := []struct {
		name        string
		target      ISCSITarget
		cmdErr      error
		expectedErr string
	}{
		{
			name: "test valid target",
			target: ISCSITarget{
				Portal: "1.1.1.1",
				Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001",
			},
			cmdErr:      nil,
			expectedErr: "",
		},
		{
			name: "test invalid portal address",
			target: ISCSITarget{
				Portal: "",
				Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001",
			},
			cmdErr:      nil,
			expectedErr: "error invalid IP or portal address",
		},
		{
			name: "test invalid IQN target",
			target: ISCSITarget{
				Portal: "1.1.1.1",
				Target: "",
			},
			cmdErr:      nil,
			expectedErr: "error invalid IQN",
		},
		{
			name: "test iscsiadm command error",
			target: ISCSITarget{
				Portal: "1.1.1.1",
				Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001",
			},
			cmdErr:      fmt.Errorf("command failed"),
			expectedErr: "command failed",
		},
		{
			name: "test iscsiadm no objects exit code",
			target: ISCSITarget{
				Portal: "1.1.1.1",
				Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001",
			},
			cmdErr:      simulateExitCode(21),
			expectedErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return nil, test.cmdErr
			}
			err := iscsi.DeleteNode(test.target)
			if err != nil {
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			} else if test.expectedErr != "" {
				t.Errorf("Expected error: %v, but got no error", test.expectedErr)
			}
		})
	}
}

func TestSetCHAPCredentials(t *testing.T) {
	reset()
	c := NewLinuxISCSI(map[string]string{})
	tgt := ISCSITarget{
		Portal: "10.0.0.0",
		Target: "iqn.1991-05.com.emc:dummyExample",
	}
	username := "username"
	chapSecret := "secret"
	err := c.SetCHAPCredentials(tgt, username, chapSecret)
	expectedError := errors.New("exec: \"iscsiadm\": executable file not found in $PATH")
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error: %v, but got: %v", expectedError, err)
	}
}

func TestMockDiscoverTargets(t *testing.T) {
	reset()
	var c ISCSIinterface
	opts := map[string]string{}
	expected := 5
	opts[MockNumberOfTargets] = fmt.Sprintf("%d", expected)
	c = NewMockISCSI(opts)
	// c = mock
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
	data, err := os.ReadFile("testdata/session_info_valid")
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
			compareStr(t, string(session.IfaceTransport), string(ISCSITransportNameTCP))
			compareStr(t, session.IfaceInitiatorname, "iqn.1994-05.com.redhat:650e84b584d")
			compareStr(t, session.IfaceIPaddress, "1.1.1.1")
			compareStr(t, string(session.ISCSISessionState), string(ISCSISessionStateLOGGEDIN))
			compareStr(t, string(session.ISCSIConnectionState), string(ISCSIConnectionStateLOGGEDIN))
			compareStr(t, session.Username, "admin")
			compareStr(t, session.Password, "foobar")
			compareStr(t, session.UsernameIn, "")
			compareStr(t, session.PasswordIn, "")
		} else {
			compareStr(t, session.Target, "iqn.2015-10.com.dell:dellemc-foobar-123-b-61ecc53a")
			compareStr(t, session.Portal, "192.168.1.2:3260")
			compareStr(t, session.SID, "13")
			compareStr(t, string(session.IfaceTransport), string(ISCSITransportNameTCP))
			compareStr(t, session.IfaceInitiatorname, "iqn.1994-05.com.redhat:650e84b585d")
			compareStr(t, session.IfaceIPaddress, "1.1.1.1")
			compareStr(t, string(session.ISCSISessionState), string(ISCSISessionStateFAILED))
			compareStr(t, string(session.ISCSIConnectionState), string(ISCSIConnectionStateFREE))
			compareStr(t, session.Username, "")
			compareStr(t, session.Password, "")
			compareStr(t, session.UsernameIn, "")
			compareStr(t, session.PasswordIn, "")
		}
	}

	// test invalid data parsing
	data, err = os.ReadFile("testdata/session_info_invalid")
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
	data, err := os.ReadFile("testdata/node_info_valid")
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
	data, err = os.ReadFile("testdata/node_info_invalid")
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

func TestGetInterfaceForTargetIP(t *testing.T) {
	tests := []struct {
		name    string
		address []string
		cmdOut  []byte
		cmdErr  error
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "no addresses",
			address: []string{},
			cmdOut:  []byte("iface0 tcp,<empty>,<empty>,lo,<empty>"),
			cmdErr:  nil,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "single address",
			address: []string{"127.0.0.1"},
			cmdOut:  []byte("iface0 tcp,<empty>,<empty>,lo,<empty>"),
			cmdErr:  nil,
			want:    map[string]string{"127.0.0.1": "iface0"},
			wantErr: false,
		},
		{
			name:    "multiple addresses",
			address: []string{"127.0.0.1", "255.0.0.1"},
			cmdOut:  []byte("iface0 tcp,<empty>,<empty>,lo,<empty>"),
			cmdErr:  nil,
			want:    map[string]string{"127.0.0.1": "iface0"},
			wantErr: false,
		},
		{
			name:    "error getting interfaces",
			address: []string{"1.2.3.4"},
			cmdOut:  []byte(""),
			cmdErr:  fmt.Errorf("iscsiadm error"),
			want:    map[string]string{},
			wantErr: true,
		},
		{
			name:    "invalid network interfaces",
			address: []string{"1.2.3.4"},
			cmdOut:  []byte("iface0 tcp,<empty>,<empty>,abc,<empty>"),
			cmdErr:  nil,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "invalid IPs",
			address: []string{"1.2.3"},
			cmdOut:  []byte("iface0 tcp,<empty>,<empty>,lo,<empty>"),
			cmdErr:  nil,
			want:    map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of LinuxISCSI
			reset()
			iscsi := NewLinuxISCSI(map[string]string{})

			// Set up mocks or other dependencies
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return tt.cmdOut, tt.cmdErr
			}

			// Call the GetInterfaceForTargetIP function
			got, err := iscsi.GetInterfaceForTargetIP(tt.address...)

			// Check if the error matches the expected error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInterfaceForTargetIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if the result matches the expected result
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInterfaceForTargetIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOptions(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name     string
		input    *ISCSIType
		expected map[string]string
	}{
		{
			name: "Non-empty options",
			input: &ISCSIType{
				options: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Empty options",
			input: &ISCSIType{
				options: map[string]string{},
			},
			expected: map[string]string{},
		},
		{
			name:     "Nil options",
			input:    &ISCSIType{},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.getOptions()
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestPerformLogin(t *testing.T) {
	iscsi := &LinuxISCSI{}

	tests := []struct {
		name        string
		target      ISCSITarget
		cmdOut      []byte
		cmdErr      error
		expectedErr string
	}{
		{
			name:        "test invalid portal address",
			target:      ISCSITarget{Portal: "invalid", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "error invalid IP or portal address",
		},
		{
			name:        "test invalid IQN target",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "invalid"},
			cmdOut:      nil,
			cmdErr:      fmt.Errorf("error invalid IQN Target invalid: invalid IQN"),
			expectedErr: "error invalid IQN",
		},
		{
			name:        "test iscsiadm executable not found",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      fmt.Errorf("exec: \"iscsiadm\": executable file not found in $PATH"),
			expectedErr: "exec: \"iscsiadm\": executable file not found in $PATH",
		},
		{
			name:        "test iscsiadm login failure",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      simulateExitCode(1),
			expectedErr: "exit status 1",
		},
		{
			name:        "test session already exists",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      simulateExitCode(15),
			expectedErr: "",
		},
		{
			name:        "test successful login",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return test.cmdOut, test.cmdErr
			}
			err := iscsi.performLogin(test.target)
			if err != nil {
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			} else {
				if test.expectedErr != "" {
					t.Errorf("Expected error: %v, but got no error", test.expectedErr)
				}
			}
		})
	}
}

func simulateExitCode(exitCode int) error {
	// Sanitize the input to ensure it's a valid exit code
	sanitizedExitCode, err := sanitizeInput(exitCode)
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("exit %d", sanitizedExitCode))
	return cmd.Run()
}

func sanitizeInput(exitCode int) (int, error) {
	// Ensure the exit code is within the range of valid exit codes: 0-255 for Unix-like systems
	if exitCode < 0 || exitCode > 255 {
		return 0, errors.New("invalid exit code: must be within 0-255")
	}
	return exitCode, nil
}

func TestPerformLogout(t *testing.T) {
	iscsi := &LinuxISCSI{}

	tests := []struct {
		name        string
		target      ISCSITarget
		cmdOut      []byte
		cmdErr      error
		expectedErr string
	}{
		{
			name:        "test invalid portal address",
			target:      ISCSITarget{Portal: "invalid", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "error invalid IP or portal address",
		},
		{
			name:        "test invalid IQN target",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "invalid"},
			cmdOut:      nil,
			cmdErr:      fmt.Errorf("error invalid IQN Target invalid: invalid IQN"),
			expectedErr: "error invalid IQN",
		},
		{
			name:        "test iscsiadm executable not found",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      fmt.Errorf("exec: \"iscsiadm\": executable file not found in $PATH"),
			expectedErr: "exec: \"iscsiadm\": executable file not found in $PATH",
		},
		{
			name:        "test iscsiadm logout failure",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      simulateExitCode(1),
			expectedErr: "exit status 1",
		},
		{
			name:        "test session already logged out",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      simulateExitCode(15),
			expectedErr: "",
		},
		{
			name:        "test successful logout",
			target:      ISCSITarget{Portal: "1.1.1.1", Target: "iqn.1992-04.com.emc:600009700bcbb70e3287017400000001"},
			cmdOut:      nil,
			cmdErr:      nil,
			expectedErr: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runCommand = func(_ *exec.Cmd) ([]byte, error) {
				return test.cmdOut, test.cmdErr
			}
			err := iscsi.performLogout(test.target)
			if err != nil {
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			} else {
				if test.expectedErr != "" {
					t.Errorf("Expected error: %v, but got no error", test.expectedErr)
				}
			}
		})
	}
}
