package infra

import (
	"encoding/json"
	"net/http"
)

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	errDbIIBB := CheckDB()
	var sDbIIBB string
	if errDbIIBB != nil {
		sDbIIBB = errDbIIBB.Error()
	} else {
		sDbIIBB = "DB CONNECTION SUCCESFULL"
	}
	if errDbIIBB != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(sDbIIBB)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(sDbIIBB)
	}
	return
}
