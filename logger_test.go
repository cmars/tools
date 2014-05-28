package httplogger_test

import (
	"bytes"
	"fmt"
	"github.com/koofr/go-httplogger"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

var expectedLog = "POST / HTTP/1.1\r\n" +
	"Host: localhost:8083\r\n" +
	"User-Agent: Go 1.1 package http\r\n" +
	"Content-Length: 3\r\n" +
	"Content-Type: text/plain\r\n" +
	"Accept-Encoding: gzip\r\n" +
	"\r\n" +
	"123\n" +
	"HTTP/1.1 200 OK\r\n" +
	"Content-Length: 14\r\n" +
	"Content-Type: text/plain; charset=utf-8\r\n" +
	"Date: Wed, 28 May 2014 12:00:27 GMT\r\n" +
	"\r\n" +
	"Hello, client\n" +
	"\n"

func TestTransportLogger(t *testing.T) {
	go http.ListenAndServe(":8083", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadAll(r.Body)

		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(content, []byte("123")) {
			t.Error("Request failed")
		}

		w.Header().Set("Date", "Wed, 28 May 2014 12:00:27 GMT")
		fmt.Fprintln(w, "Hello, client")
	}))

	buffer := new(bytes.Buffer)

	logger := httplogger.New(http.DefaultTransport)
	logger.Writer = buffer

	client := &http.Client{
		Transport: logger,
	}

	res, err := client.Post("http://localhost:8083", "text/plain", bytes.NewReader([]byte("123")))

	if err != nil {
		t.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)

	res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(greeting, []byte("Hello, client\n")) {
		t.Error("Response failed")
	}

	if !reflect.DeepEqual(buffer.Bytes(), []byte(expectedLog)) {
		t.Error("Logger failed")
	}
}
