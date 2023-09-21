package origin

import "errors"

var (
	ErrNilResponse   = errors.New("nil http response")
	ErrNilHttpClient = errors.New("nil http client")
	ErrInvalidStatus = errors.New("invalid http status of reponse")
)
