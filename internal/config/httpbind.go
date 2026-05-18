package config

import (
	"errors"
	"net/netip"
	"strconv"
	"strings"
)

// ValidateHTTPBindAddress rejects accidental wildcard binds and malformed ports.
func ValidateHTTPBindAddress(value string) error {
	_, err := NormalizeHTTPBindAddress(value)
	return err
}

// NormalizeHTTPBindAddress validates and returns value as a canonical netip.AddrPort string.
func NormalizeHTTPBindAddress(value string) (string, error) {
	host, port, err := splitHTTPBindAddress(value)
	if err != nil {
		return "", err
	}
	if host == "" {
		return "", errors.New("invalid HTTP bind address; use an explicit IP host and port like 127.0.0.1:8765")
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return "", errors.New("invalid HTTP bind address; port must be between 1 and 65535")
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return "", errors.New("invalid HTTP bind address; host must be an IP address like 127.0.0.1 or 192.168.1.10")
	}
	if !addr.IsValid() {
		return "", errors.New("invalid HTTP bind address; host must be a valid IP address")
	}
	return netip.AddrPortFrom(addr, uint16(portNumber)).String(), nil
}

// HTTPBindAddressIsLoopback reports whether value binds only to a loopback IP.
func HTTPBindAddressIsLoopback(value string) bool {
	normalized, err := NormalizeHTTPBindAddress(value)
	if err != nil {
		return false
	}
	addrPort, err := netip.ParseAddrPort(normalized)
	return err == nil && addrPort.Addr().IsLoopback()
}

func splitHTTPBindAddress(value string) (string, string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", "", errors.New("invalid HTTP bind address; use host:port like 127.0.0.1:8765")
	}
	host, port, ok := strings.Cut(trimmed, ":")
	if !ok || strings.Contains(port, ":") || strings.Contains(host, ":") {
		addrPort, err := netip.ParseAddrPort(trimmed)
		if err != nil {
			return "", "", errors.New("invalid HTTP bind address; use host:port like 127.0.0.1:8765 or [::1]:8765")
		}
		return addrPort.Addr().String(), strconv.Itoa(int(addrPort.Port())), nil
	}
	return strings.TrimSpace(host), strings.TrimSpace(port), nil
}
