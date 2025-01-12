package main

import (
	"context"
	"net"

	"github.com/xocoder/hysteria/pkg/utils"
)

func setResolver(dns string) {
	if _, _, err := utils.SplitHostPort(dns); err != nil {
		// Append the default DNS port
		dns = net.JoinHostPort(dns, "53")
	}
	dialer := net.Dialer{}
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, dns)
		},
	}
}
