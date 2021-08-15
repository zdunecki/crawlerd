package util

// TODO: util should not be in pkg

import (
	"errors"
	"net"
	"net/url"
)

var privateIPNetworks = []net.IPNet{
	{
		IP:   net.ParseIP("10.0.0.0"),
		Mask: net.CIDRMask(8, 32),
	},
	{
		IP:   net.ParseIP("172.16.0.0"),
		Mask: net.CIDRMask(12, 32),
	},
	{
		IP:   net.ParseIP("192.168.0.0"),
		Mask: net.CIDRMask(16, 32),
	},
}

func ResolveHostIP() (string, error) {
	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {

		networkIp, ok := netInterfaceAddress.(*net.IPNet)

		if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {

			ip := networkIp.IP.String()

			return ip, nil
		}
	}

	return "", errors.New("empty IP")
}

func BaseAddr(s string) string {
	u, _ := url.Parse(s)
	port := u.Port()

	host := ""

	if port != "" {
		host = net.JoinHostPort(u.Hostname(), u.Port())
	} else {
		host = u.Hostname()
	}

	newU := url.URL{
		Scheme: u.Scheme,
		Host:   host,
	}

	return newU.String()
}
