package push

import (
	"time"

	"github.com/doyensec/safeurl"
)

// NewSafeClient returns an HTTP client with SSRF protection
// Set allowLocalhost=true to permit 127.0.0.1 for local development.
func NewSafeClient(allowLocalhost bool) *safeurl.WrappedClient {
	builder := safeurl.GetConfigBuilder().
		SetTimeout(30*time.Second).
		SetAllowedSchemes("http", "https").
		SetAllowedPorts(80, 443, 8000, 8080, 8443, 3000, 5000). // common API ports
		EnableIPv6(false)                                       // disable IPv6 for now (increases complexity attack surface)

	if allowLocalhost {
		builder.SetAllowedIPs("127.0.0.1")
	}

	return safeurl.Client(builder.Build())
}

// NewSafeClientWithPort creates a client allowing a specific port (for testing)
func NewSafeClientWithPort(allowLocalhost bool, port int) *safeurl.WrappedClient {
	builder := safeurl.GetConfigBuilder().
		SetTimeout(30*time.Second).
		SetAllowedSchemes("http", "https").
		SetAllowedPorts(port).
		EnableIPv6(false) // disable IPv6 for now (increases complexity attack surface)

	if allowLocalhost {
		builder.SetAllowedIPs("127.0.0.1")
	}

	return safeurl.Client(builder.Build())
}
