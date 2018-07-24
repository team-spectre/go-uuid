package uuid

import (
	"fmt"
)

// Version indicates the RFC 4122-defined UUID version.
type Version byte

// Version enum constants.
const (
	V1 Version = 1
	V2 Version = 2
	V3 Version = 3
	V4 Version = 4
	V5 Version = 5
)

var versionMap = map[Version]string{
	V1: "V1",
	V2: "V2",
	V3: "V3",
	V4: "V4",
	V5: "V5",
}

// IsValid returns true iff the Version is one of the UUID versions known to RFC 4122.
func (version Version) IsValid() bool {
	return (version >= V1 && version <= V5)
}

func (version Version) String() string {
	if str, found := versionMap[version]; found {
		return str
	}
	return fmt.Sprintf("V%d", version)
}

// FromString attempts to parse the string representation of a Version.
func (version *Version) FromString(in string) error {
	for k, v := range versionMap {
		if v == in {
			*version = k
			return nil
		}
	}
	var b byte
	n, err := fmt.Sscanf(in+"|", "V%d|", &b)
	if n == 1 && err == nil {
		*version = Version(b)
		return nil
	}
	return makeParseError("Version", "FromString", []byte(in), true)
}
