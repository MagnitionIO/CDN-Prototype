package origin

import (
	"io"
	"net/http"
	"unsafe"
)

type StringResponse struct {
	*http.Response
}

func (r *StringResponse) Result() (string, error) {
	if r.Response == nil {
		return "", ErrNilResponse
	}

	if r.Response.StatusCode != http.StatusOK {
		return "", ErrInvalidStatus
	}

	s, err := io.ReadAll(r.Response.Body)
	if err != nil {
		return "", err
	}

	return unsafe.String(unsafe.SliceData(s), len(s)), nil

}

type ObjectResponse struct {
	StringResponse
	ID   int
	Size int
}
