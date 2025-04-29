package app

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func isValidSocket(socket string) error {
	s := strings.SplitN(socket, ":", 2)
	if len(s) < 2 {
		return fmt.Errorf("invalid format, the valid format: <protocol>:<socket-addr>")
	}
	switch strings.ToLower(s[0]) {
	case UNIX:
		if len(s[1]) > 0 {
			return nil
		}
		return fmt.Errorf("invalid unix domain socket")
	case TCP:
		host, port, err := net.SplitHostPort(s[1])
		if err != nil {
			return fmt.Errorf("invalid tcp socket: %w", err)
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("invalid ip address: %s", host)
		}
		if !isValidPort(port) {
			return fmt.Errorf("invalid port: %s", port)
		}
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s, supported values: tcp and unix", s[0])
	}
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= 0 && p <= 65535
}

func validatePortRange(ranges []string) error {
	rangePattern := regexp.MustCompile(`^(\d+)(?:-(\d+))?$`)
	for _, r := range ranges {
		matches := rangePattern.FindStringSubmatch(r)
		if matches == nil {
			return fmt.Errorf("invalid port format: %s", r)
		}
		start, _ := strconv.Atoi(matches[1])
		if !isValidPort(matches[1]) {
			return fmt.Errorf("invalid start port: %s", matches[1])
		}
		if matches[2] != "" {
			end, _ := strconv.Atoi(matches[2])
			if !isValidPort(matches[2]) {
				return fmt.Errorf("invalid end port: %s", matches[2])
			}
			if end < start {
				return fmt.Errorf("invalid range: %s (end < start)", r)
			}
		}
	}
	return nil
}

func validateAddressRange(addresses []string) error {
	for _, addr := range addresses {
		if strings.Contains(addr, "/") {
			if _, _, err := net.ParseCIDR(addr); err != nil {
				return fmt.Errorf("invalid CIDR: %s", addr)
			}
		} else {
			if ip := net.ParseIP(addr); ip == nil {
				return fmt.Errorf("invalid IP address: %s", addr)
			}
		}
	}
	return nil
}

// Checks if the given port (int) is within any of the specified port ranges
func isPortInRange(portStr string, ranges []string) bool {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return false
	}
	for _, r := range ranges {
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && port >= start && port <= end {
				return true
			}
		} else {
			p, err := strconv.Atoi(r)
			if err == nil && port == p {
				return true
			}
		}
	}
	return false
}

// Checks if the given IP string belongs to any of the CIDRs or IPs in addressRange
func isIPInRange(ipStr string, ranges []string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, r := range ranges {
		if strings.Contains(r, "/") {
			_, cidrNet, err := net.ParseCIDR(r)
			if err == nil && cidrNet.Contains(ip) {
				return true
			}
		} else {
			if rIP := net.ParseIP(r); rIP != nil && rIP.Equal(ip) {
				return true
			}
		}
	}
	return false
}
