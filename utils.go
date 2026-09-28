/*
 *
 * Copyright © 2022-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"net"
	"regexp"
	"strconv"

	"github.com/dell/csmlog"
)

func validateIPAddress(ip string) error {
	if net.ParseIP(ip) != nil {
		return nil
	}

	// net.SplitHostPort accepts IPv4 host:port and bracketed IPv6 [host]:port.
	host, port, err := net.SplitHostPort(ip)
	if err == nil && net.ParseIP(host) != nil {
		if port == "" {
			return errors.New("error invalid IP or portal address: empty port")
		}
		portNum, err := strconv.ParseUint(port, 10, 32)
		if err != nil {
			return errors.New("error invalid IP or portal address: non-numeric port")
		}
		if portNum < 1 || portNum > 65535 {
			return errors.New("error invalid IP or portal address: port out of range")
		}
		return nil
	}

	return errors.New("error invalid IP or portal address")
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
		csmlog.WithFields(csmlog.Fields{
			csmlog.FieldComponent: "goiscsi",
			csmlog.FieldOperation: "filterIPsForInterface",
			csmlog.FieldError:     err.Error(),
			"interface":           ifaceName,
		}).Error("could not find interface")
		return filteredIPs, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		csmlog.WithFields(csmlog.Fields{
			csmlog.FieldComponent: "goiscsi",
			csmlog.FieldOperation: "filterIPsForInterface",
			csmlog.FieldError:     err.Error(),
			"interface":           ifaceName,
		}).Error("failed to get addresses of interface")
		return filteredIPs, err
	}

	for _, ipAddr := range ipAddress {

		ip := net.ParseIP(ipAddr)
		if ip == nil {
			csmlog.WithFields(csmlog.Fields{
				csmlog.FieldComponent: "goiscsi",
				csmlog.FieldOperation: "filterIPsForInterface",
				"ip_address":          ipAddr,
			}).Error("invalid IP address")
			continue
		}
		for _, addr := range addrs {
			ifaceIP, ifaceSubnet, err := net.ParseCIDR(addr.String())
			if err != nil {
				csmlog.WithFields(csmlog.Fields{
					csmlog.FieldComponent: "goiscsi",
					csmlog.FieldOperation: "filterIPsForInterface",
					csmlog.FieldError:     err.Error(),
					"address":             addr.String(),
					"interface":           ifaceName,
				}).Error("failed to parse address")
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
