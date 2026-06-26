package mcgo

import (
	"encoding/json"
	"io"
)

// jsonDecode parse JSON desde un reader.
func jsonDecode(r io.Reader, v interface{}) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(v)
}
