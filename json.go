package ginflux

import (
	jn "github.com/json-iterator/go"
)

var json jn.API

func init() {
	//extra.RegisterFuzzyDecoders()
	//extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	json = jn.ConfigFastest
}
