package types

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// GenericBackendResponse is the generic format for the backend response.
type GenericBackendResponse map[string]any

func NewGenericBackendResponseFrom(v any) (res GenericBackendResponse, err error) {
	var bz []byte
	bz, err = json.Marshal(v)
	if err != nil {
		err = errors.Wrap(err, "failed to marshal input")
		return
	}

	err = json.Unmarshal(bz, &res)
	if err != nil {
		err = errors.Wrap(err, "failed to build response")
		return
	}

	return
}
