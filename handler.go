package pjax

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"code.google.com/p/go-html-transform/css/selector"
	"code.google.com/p/go-html-transform/h5"
)

type PjaxResultRewriter struct {
	handler http.HandlerFunc
}

func NewPjaxResultRewriter(handler http.HandlerFunc) *PjaxResultRewriter {
	return &PjaxResultRewriter{
		handler: handler,
	}
}

func rewriteBody(containerSelector string, dest io.Writer, body string) (err error) {
	if containerSelector == "" {
		dest.Write([]byte(body))
		return
	}

	var chain *selector.Chain
	var document *h5.Tree

	if document, err = h5.NewFromString(body); err != nil {
		err = fmt.Errorf("invalid html document: %v", err)
		return
	}

	if chain, err = selector.Selector(containerSelector); err != nil {
		err = fmt.Errorf("invalid css: %v", containerSelector)
		return
	}

	if matches := chain.Find(document.Top()); len(matches) > 0 {
		match := matches[0:1] // Take only the first match
		newBody := h5.RenderNodesToString(match)

		dest.Write([]byte(newBody))
		return
	}

	err = fmt.Errorf("container not found")
	return
}

func (s *PjaxResultRewriter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var container string

	// Validate pjax meta data.
	if container = req.Header.Get("X-PJAX-CONTAINER"); container == "" {
		container = req.URL.Query().Get("_pjax")
	}

	// We didn't find a container, so handle as non pjax request.
	if container == "" {
		// Request contains no pjax information.
		s.handler(rw, req)
		return
	}

	response := NewInMemoryResponseWriter()
	s.handler(response, req)

	for key, _ := range response.Header() {
		if key != "Content-Lenght" {
			value := response.Header().Get(key)
			response.Header().Set(key, value)
		}
	}

	if rewriteErr := rewriteBody(container, rw, response.body.String()); rewriteErr != nil {
		response.body.WriteTo(rw)
	}
}

type InMemoryResponseWriter struct {
	writeHeaderCalled bool
	header            http.Header
	body              bytes.Buffer
}

func NewInMemoryResponseWriter() *InMemoryResponseWriter {
	return &InMemoryResponseWriter{
		header: make(http.Header),
	}
}

// Header returns the header map that will be sent by WriteHeader.
// Changing the header after a call to WriteHeader (or Write) has
// no effect.
func (w *InMemoryResponseWriter) Header() http.Header {
	return w.header
}

// Write writes the data to the connection as part of an HTTP reply.
// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
// before writing the data.  If the Header does not contain a
// Content-Type line, Write adds a Content-Type set to the result of passing
// the initial 512 bytes of written data to DetectContentType.
func (w *InMemoryResponseWriter) Write(p []byte) (int, error) {
	if !w.writeHeaderCalled {
		w.WriteHeader(http.StatusOK)
	}

	if w.header.Get("Content-Type") == "" {
		w.header.Set("Content-Type", http.DetectContentType(p))
	}

	return w.body.Write(p)
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (w *InMemoryResponseWriter) WriteHeader(code int) {
	w.writeHeaderCalled = true
	w.header.Set("Status-Code", strconv.Itoa(code))
}

func (w *InMemoryResponseWriter) WriteTo(writer io.Writer) error {
	if err := w.header.Write(writer); err != nil {
		return err
	}
	if _, err := w.body.WriteTo(writer); err != nil {
		return err
	}

	return nil
}
