package function

import (
	"fmt"
	"net/http"
)

func String() string {
	return "utils"
}

// NewService creates a new http handler
func NewService() interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		input := String()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hello world, input was: %s", string(input))))
	})
}
