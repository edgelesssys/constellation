package version

const (
	Bootloader Type = "bootloader" // Bootloader is the version of the Azure SEVSNP bootloader.
	TEE        Type = "tee"        // TEE is the version of the Azure SEVSNP TEE.
	SNP        Type = "snp"        // SNP is the version of the Azure SEVSNP SNP.
	Microcode  Type = "microcode"  // Microcode is the version of the Azure SEVSNP microcode.
)

// Type is the type of the version to be requested.
type Type (string)

// Version stores the version of a given type.
//type Version (string)

//// NewVersion validates that the given string is either "latest" or uint version number.
//func NewVersion(raw string) (Version, error) {
//	if raw != "latest" {
//		_, err := strconv.ParseUint(raw, 10, 8)
//		if err != nil {
//			return Version(raw), err
//		}
//	}
//	return Version(raw), nil
//}

//// Value returns the uint value of the version.
//func (v Version) Value() uint8 {
//	// ignore error as it is already validated in NewVersion
//	res, _ := strconv.ParseUint(string(v), 10, 8)
//	return uint8(res)
//}

//// UnmarshalYAML implements a custom unmarshaler to resolve the latest version value.
//func (v *Version) UnmarshalYAML(unmarshal func(interface{}) error) error {
//	var raw string
//	if err := unmarshal(&raw); err != nil {
//		return err
//	}
//	res, err := NewVersion(raw)
//	if err != nil {
//		return fmt.Errorf("invalid version %q: %w", raw, err)
//	}
//	*v = res
//	return nil
//}

// GetVersion returns the version of the given type.
func GetVersion(t Type) uint8 {
	switch t {
	case Bootloader:
		return 2
	case TEE:
		return 0
	case SNP:
		return 6
	case Microcode:
		return 93
	default:
		return 1
	}
}
