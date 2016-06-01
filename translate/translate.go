/*
Package translate handles translation between configuration
specifications.

For example, it allows you to generate OCI-compliant config.json from
a higher-level configuration language.
*/
package translate

import (
	"github.com/codegangsta/cli"
)

// Translate maps JSON from one specification to another.
type Translate func(data interface{}, context *cli.Context) (translated interface{}, err error)

// Translators is a map from translator names to Translate functions.
var Translators = map[string]Translate{
	"fromContainer": FromContainer,
}
