package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const maxBodySize = 1 << 20 // 1 MB

// decodeJSON reads and decodes the request body into dst.
// It enforces a 1MB size limit and rejects unknown fields.
func decodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}
	return nil
}
