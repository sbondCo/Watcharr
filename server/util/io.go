package util

import (
	"errors"
	"io"
)

func LimitedReadAll(r io.Reader, maxSize int64) ([]byte, error) {
	// LimitReader: Allow reading one extra byte so our maxSize check later works.
	b, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil {
		return b, err
	}
	if int64(len(b)) > maxSize {
		return b, errors.New("file is too big")
	}
	return b, nil
}
