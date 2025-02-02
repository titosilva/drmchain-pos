package errorutil

import "errors"

func WithInner(errMsg string, inner error) error {
	err := errors.New(errMsg)
	return errors.Join(err, inner)
}
