package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/example/creator-loop-observer/deliveryloop"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	service := deliveryloop.NewService(apiKey)

	http.HandleFunc("/deliveries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		var input deliveryloop.DeliveryRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid delivery request", http.StatusBadRequest)
			return
		}
		result, err := service.Process(r.Context(), input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	log.Println("creator delivery loop listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
