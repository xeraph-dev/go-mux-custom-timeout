package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func timeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		var cancel context.CancelFunc
		ctx = context.Background()
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		done := make(chan struct{})
		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()

		select {
		case <-done:
			println("all good")
		case <-ctx.Done():
			switch err := ctx.Err(); err {
			case context.DeadlineExceeded:
				w.WriteHeader(http.StatusAccepted)
				io.WriteString(w, "processing")
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	})
}

func handler(w http.ResponseWriter, r *http.Request) {

}

func slowHandlerThatPass(w http.ResponseWriter, r *http.Request) {
	println("starting handler")
	time.Sleep(4 * time.Second)
	println("finish first slow operation")
	// time.Sleep(4 * time.Second)
	println("ending handler")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "finished")
}

func slowHandlerNotPass(w http.ResponseWriter, r *http.Request) {
	println("starting handler")
	time.Sleep(4 * time.Second)
	println("finish first slow operation")
	time.Sleep(4 * time.Second)
	println("ending handler")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "finished")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	r.Handle("/slow-pass", timeoutMiddleware(http.HandlerFunc(slowHandlerThatPass)))
	r.Handle("/slow-not-pass", timeoutMiddleware(http.HandlerFunc(slowHandlerNotPass)))
	http.ListenAndServe(":8080", r)
}
