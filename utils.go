/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"net"
	"regexp"
)

func validateIPAddress(ip string) error {
	isValidIP := true
	isValidPortal := true

	// validtes only IP
	if net.ParseIP(ip) == nil {
		isValidIP = false
	}

	// Regex to validate IPV4 with port - for portal validation Ex: 10.0.0.0:1111
	const exp = `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]):[0-9]+$`
	r := regexp.MustCompile(exp)
	if !r.MatchString(ip) {
		isValidPortal = false
	}
	// Either valid IP/portal address should be given
	if !isValidIP && !isValidPortal {
		return errors.New("error invalid IP or portal address")
	}
	return nil
}

func validateIQN(iqn string) error {
	const exp = `iqn\.\d{4}-\d{2}\.([[:alnum:]-.]+)(:[^,;*&$|\s]+)$`
	r := regexp.MustCompile(exp)
	if !r.MatchString(iqn) {
		return errors.New("error invalid IQN")
	}
	return nil
}

func filterIPsForInterface(ifaceName string, ipAddress ...string) ([]string, error) {
	filteredIPs := make([]string, 0)
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		fmt.Printf("\nError could not find interface %s : %v", ifaceName, err)
		return filteredIPs, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Printf("\nError failed to get addresses of interface %s : %v", ifaceName, err)
		return filteredIPs, err
	}

	for _, ipAddr := range ipAddress {

		ip := net.ParseIP(ipAddr)
		if ip == nil {
			fmt.Printf("\nError invalid IP address: %s", ipAddr)
			continue
		}
		for _, addr := range addrs {
			ifaceIP, ifaceSubnet, err := net.ParseCIDR(addr.String())
			if err != nil {
				fmt.Printf("\nError failed to parse address %s of interface %s : %v", addr.String(), ifaceName, err)
				continue
			}

			// Check if the IP belongs to the subnet
			if ifaceSubnet.Contains(ip) || ifaceIP.Equal(ip) {
				filteredIPs = append(filteredIPs, ipAddr)
				break
			}
		}

	}

	return filteredIPs, nil
}
