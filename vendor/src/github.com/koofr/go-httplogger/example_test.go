package httplogger_test

import (
	"bytes"
	"fmt"
	"github.com/koofr/go-httplogger"
	"io/ioutil"
	"log"
	"net/http"
)

func ExampleTransportLogger() {
	go http.ListenAndServe(":8083", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "Wed, 28 May 2014 12:00:27 GMT")
		fmt.Fprintln(w, "Hello, client")
	}))

	client := &http.Client{
		Transport: httplogger.New(http.DefaultTransport),
	}

	res, err := client.Post("http://localhost:8083", "text/plain", bytes.NewReader([]byte("123")))

	if err != nil {
		log.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)

	res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", greeting)
}
