package pjax

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
)

type ResponseRecorder struct {
	writeHeaderCalled bool
	header            http.Header
	body              bytes.Buffer
}

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		header: make(http.Header),
	}
}

// Header returns the header map that will be sent by WriteHeader.
// Changing the header after a call to WriteHeader (or Write) has
// no effect.
func (w *ResponseRecorder) Header() http.Header {
	return w.header
}

func (w *ResponseRecorder) Write(p []byte) (int, error) {
	if !w.writeHeaderCalled {
		w.WriteHeader(http.StatusOK)
	}

	return w.body.Write(p)
}

func (w *ResponseRecorder) WriteHeader(code int) {
	w.writeHeaderCalled = true
	w.header.Set("Status-Code", strconv.Itoa(code))
}

func (w *ResponseRecorder) WriteTo(writer io.Writer) error {
	if err := w.header.Write(writer); err != nil {
		return err
	}
	if _, err := w.body.WriteTo(writer); err != nil {
		return err
	}

	return nil
}
