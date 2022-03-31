package cmd

const (
	// wireguardAdminMTU is the MTU designated for the admin's WireGuard interface.
	//
	// WireGuard doesn't support Path MTU Discovery. Thus, its default MTU can be too high on some networks.
	wireguardAdminMTU = 1300

	// masterSecretLengthDefault is the default length in bytes for CLI generated master secrets.
	masterSecretLengthDefault = 32

	// masterSecretLengthMin is the minimal length in bytes for user provided master secrets.
	masterSecretLengthMin = 16

	// constellationNameLength is the maximum length of a Constellation's name.
	constellationNameLength = 37
)
