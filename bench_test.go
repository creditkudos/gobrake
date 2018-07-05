package gobrake_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/airbrake/gobrake"
)

func BenchmarkSendNotice(b *testing.B) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))

	notifier := gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:  1,
		ProjectKey: "key",
		Host:       server.URL,
	})

	notice := notifier.Notice(errors.New("benchmark"), nil, 0)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id, err := notifier.SendNotice(notice)
			if err != nil {
				b.Fatal(err)
			}
			if id != "123" {
				b.Fatalf("got %q, wanted 123", id)
			}
		}
	})
}
