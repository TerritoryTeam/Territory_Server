package api

import (
	"encoding/json"
	"net/http"
)

type TestResponse struct {
	Success bool `json:"success"`
}

func HttpTest(w http.ResponseWriter, r *http.Request) {
	response := &TestResponse{Success: true}

	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
