package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func ReadNotDrain(r *http.Request) (content []byte, err error) {
	content, err = ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(content))
	return
}
