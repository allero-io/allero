package jsonschemaValidator

import (
	"github.com/xeipuuv/gojsonschema"
)

func Validate(schema []byte, document interface{}) (*gojsonschema.Result, error) {
	loader := gojsonschema.NewStringLoader(string(schema))
	documentLoader := gojsonschema.NewGoLoader(document)

	return gojsonschema.Validate(loader, documentLoader)
}
