package handler

import "net/http"

func API() *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}
