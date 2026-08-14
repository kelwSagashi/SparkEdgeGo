package updater

import (
	"encoding/json"
	"io"
)

func decodeJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	return decoder.Decode(target)
}
