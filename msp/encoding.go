package msp

import (
	"encoding/json"
	"github.com/dottics/dutil"
	"io"
	"net/http"
)

// Decode is a function that decodes a body into a slice of bytes and also
// will unmarshal the data into an interface pointer value if the value
// pointed to by the interface is provided.
func Decode(res *http.Response, v interface{}) ([]byte, dutil.Error) {
	xb, err := io.ReadAll(res.Body)
	if err != nil {
		e := dutil.NewErr(500, "read", []string{err.Error()})
		return nil, e
	}
	err = res.Body.Close()
	if err != nil {
		e := dutil.NewErr(500, "readClose", []string{err.Error()})
		return nil, e
	}
	if v != nil {
		err = json.Unmarshal(xb, v)
		if err != nil {
			e := dutil.NewErr(500, "unmarshal", []string{err.Error()})
			return nil, e
		}
	}
	return xb, nil
}
