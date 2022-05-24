package goiscsi

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

func validateIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		fmt.Printf("IP Address: %s - Invalid\n", ip)
		return errors.New("Invalid IP address")
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
