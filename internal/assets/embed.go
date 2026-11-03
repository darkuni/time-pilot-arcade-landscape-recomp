//go:build embedrom

package assets

import _ "embed"

// Personal-use build: the ROM zip is placed here before building and baked
// into the executable.  Never distribute such a build.
//
//go:embed timeplt.zip
var embeddedZip []byte

func init() { Embedded = embeddedZip }
