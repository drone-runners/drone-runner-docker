package resource

import (
	"fmt"
	"strings"
)

// ParseVolume parses a volume string of the form <src>:<dest>
// and returns the source, destination and whether the volume
// is read only.
func ParseVolume(v string) (src, dest string, ro bool, err error) {
	plen := 2
	z := strings.SplitN(v, ":", plen)
	if len(z) != plen {
		return src, dest, ro, fmt.Errorf("volume %s is not in the format src:dest", v)
	}
	src = z[0]
	dest = z[1]
	ro = strings.HasSuffix(dest, ":ro")
	dest = strings.TrimSuffix(dest, ":ro")
	return src, dest, ro, nil
}
