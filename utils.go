package goiscsi

import (
	"errors"
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
		return errors.New("Error invalid IP or portal address")
	}
	return nil
}

func validateIQN(iqn string) error {
	const exp = `iqn\.\d{4}-\d{2}\.([[:alnum:]-.]+)(:[^,;*&$|\s]+)$`
	r := regexp.MustCompile(exp)
	if !r.MatchString(iqn) {
		return errors.New("Error invalid IQN")
	}
	return nil
}
