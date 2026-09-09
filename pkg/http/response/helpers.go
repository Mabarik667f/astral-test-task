package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, resp Resp) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func ErrResponse(w http.ResponseWriter, status int, text string) {
	JSON(w, status, Resp{
		Error: &Err{
			Code: status,
			Text: text,
		},
	})
}
