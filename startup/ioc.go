package startup

import (
	"github.com/condigolabs/content-marketing/services/generator"
	"github.com/condigolabs/content-marketing/services/intent"
)

var gen generator.Generator
var inte intent.Intent

func GetGenerator() generator.Generator {
	if gen == nil {
		gen, _ = generator.New()
	}
	return gen
}
func GetIntent() intent.Intent {
	if inte == nil {
		inte, _ = intent.New()
	}
	return inte
}

func Close() {
	if gen != nil {
		gen.Close()
	}
	if inte != nil {
		inte.Close()
	}

}
