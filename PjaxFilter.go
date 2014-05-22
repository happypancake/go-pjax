package pjax

import (
	"fmt"
	"io"
	"net/http"

	"code.google.com/p/go-html-transform/css/selector"
	"code.google.com/p/go-html-transform/h5"
	"code.google.com/p/go.net/html"
)

type PjaxFilter struct {
	handler http.HandlerFunc
}

func NewPjaxFilter(handler http.HandlerFunc) *PjaxFilter {
	return &PjaxFilter{
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

	var titleNode string
	if titleNode, err = getTitleNode(document); err != nil {
		return
	}

	if chain, err = selector.Selector(containerSelector); err != nil {
		err = fmt.Errorf("invalid css: %v", containerSelector)
		return
	}

	if matches := chain.Find(document.Top()); len(matches) > 0 {
		match := matches[0:1] // Take only the first match
		newBody := h5.RenderNodesToString([]*html.Node{match[0].FirstChild})

		fmt.Printf("data: %v", h5.Data(match[0]))

		dest.Write([]byte(titleNode))
		dest.Write([]byte(newBody))
		return
	}

	err = fmt.Errorf("container not found")
	return
}

func getTitleNode(document *h5.Tree) (titleNode string, err error) {
	var chain *selector.Chain

	if chain, err = selector.Selector("title"); err != nil {
		return
	}

	if matches := chain.Find(document.Top()); len(matches) > 0 {
		match := matches[0:1]
		titleNode = h5.RenderNodesToString(match)
	}

	return
}

func (s *PjaxFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

	response := NewResponseRecorder()
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
