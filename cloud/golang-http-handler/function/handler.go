package function

import (
	"fmt"
	"net/http"

	"function/utils"
)

// NewHandler creates a new http handler
func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		input := utils.String()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hello world, input was: %s", string(input))))
	})
}
