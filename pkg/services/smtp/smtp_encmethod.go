package smtp

import (
	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

type encMethod int

const (
	EncNone        encMethod = iota // 0
	EncExplicitTLS                  // 1
	EncImplicitTLS                  // 2
	EncAuto                         // 3
)

type encMethodVals struct {
	// None means no encryption
	None encMethod
	// ExplicitTLS means that TLS needs to be initiated by using StartTLS
	ExplicitTLS encMethod
	// ImplicitTLS means that TLS is used for the whole session
	ImplicitTLS encMethod
	// Auto means that TLS will be implicitly used for port 465, otherwise explicit TLS will be used if its supported
	Auto encMethod

	// Enum is the EnumFormatter instance for EncMethods
	Enum types.EnumFormatter
}

// EncMethods is the enum helper for populating the Encryption field.
var EncMethods = &encMethodVals{
	None:        EncNone,
	ExplicitTLS: EncExplicitTLS,
	ImplicitTLS: EncImplicitTLS,
	Auto:        EncAuto,

	Enum: format.CreateEnumFormatter(
		[]string{
			"None",
			"ExplicitTLS",
			"ImplicitTLS",
			"Auto",
		}),
}

func (at encMethod) String() string {
	return EncMethods.Enum.Print(int(at))
}

func useImplicitTLS(encryption encMethod, port uint16) bool {
	switch encryption {
	case EncMethods.ImplicitTLS:
		return true
	case EncMethods.Auto:
		return port == ImplicitTLSPort
	default:
		return false
	}
}

// ImplicitTLSPort is de facto standard SMTPS port.
const ImplicitTLSPort = 465
