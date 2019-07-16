package ginflux

import (
	. "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json API

func init() {
	extra.RegisterFuzzyDecoders()
	//extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	json = ConfigCompatibleWithStandardLibrary
}
