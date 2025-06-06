package checkapi

import (
	"context"
	"encoding/json"
	"net/http"
)

func liveness(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	status := struct {
		Status string
	}{
		Status: "ok",
	}

	json.NewEncoder(w).Encode(status)
}

func readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	status := struct {
		Status string
	}{
		Status: "ok",
	}

	json.NewEncoder(w).Encode(status)
}
