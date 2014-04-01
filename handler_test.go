package pjax

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerProxies(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)

	server := NewPjaxResultRewriter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.Write([]byte("hello world"))
	})

	server.ServeHTTP(recorder, request)

	if contentType := recorder.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Fatalf("wrong content type: ", contentType)
	}

	body, _ := ioutil.ReadAll(recorder.Body)
	assert.Equal(t, "<div id=\"main\">hello world</div>", recorder.Body.String())
}

func TestHandlerSelectContainerOnPjaxRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://foobar", nil)
	request.Header.Add("X-PJAX", "true")
	request.Header.Add("X-PJAX-CONTAINER", "#main")

	server := NewPjaxResultRewriter(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("<html><head><title>foobar</title></head><body><div id=\"main\">hello world</div></body></html>"))
	})

	server.ServeHTTP(recorder, request)

	assert.Equal(t, "<div id=\"main\">hello world</div>", recorder.Body.String())
}
