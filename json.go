package ginflux

import (
	jn "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json jn.API

func init() {
	extra.RegisterFuzzyDecoders()
	//extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	json = jn.ConfigCompatibleWithStandardLibrary
}
