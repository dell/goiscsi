package goiscsi

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

func checkIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		fmt.Printf("IP Address: %s - Invalid\n", ip)
		return errors.New("Invalid IP address")
	}
	return nil
}

func checkIQN(iqn string) error {
	r := regexp.MustCompile(`iqn\.\d{4}-\d{2}\.([[:alnum:]-.]+)(:[^,;*&$|\s]+)$`)
	if !r.MatchString(iqn) {
		return errors.New("Error invalid IQN")
	}
	return nil
}
