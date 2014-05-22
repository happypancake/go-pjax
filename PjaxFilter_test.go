package pjax

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerProxies(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)

	server := NewPjaxFilter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("hello world"))
	})

	server.ServeHTTP(recorder, request)

	if contentType := recorder.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Fatalf("wrong content type: ", contentType)
	}

	assert.Equal(t, "hello world", recorder.Body.String())
}

func TestHandlerSelectContainerOnPjaxRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)
	request.Header.Add("X-PJAX", "true")
	request.Header.Add("X-PJAX-CONTAINER", "#main")

	server := NewPjaxFilter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("<html><head><body><div id=\"main\">hello world</div></body></html>"))
	})

	server.ServeHTTP(recorder, request)

	assert.Equal(t, "hello world", recorder.Body.String())
}

func TestHandlerSelectTitleOnPjaxRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)
	request.Header.Add("X-PJAX", "true")
	request.Header.Add("X-PJAX-CONTAINER", "#main")

	server := NewPjaxFilter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("<html><head><title>foobar</title></head><body><div id=\"main\">hello world</div></body></html>"))
	})

	server.ServeHTTP(recorder, request)

	assert.True(t, strings.HasPrefix(recorder.Body.String(), "<title>foobar</title"))
}

func TestHandlerSelectsBodyWithMultipleChildrenOnPjaxRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)
	request.Header.Add("X-PJAX", "true")
	request.Header.Add("X-PJAX-CONTAINER", "#main")

	server := NewPjaxFilter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("<html><head><title>foobar</title></head><body><div id=\"main\"><h1>title</h1><p>hello world</p></div></body></html>"))
	})

	server.ServeHTTP(recorder, request)

	assert.Equal(t, recorder.Body.String(), "<title>foobar</title><h1>title</h1><p>hello world</p>")
}
