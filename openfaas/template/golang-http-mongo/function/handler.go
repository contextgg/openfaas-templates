package function

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

func NewHandler(db *mongo.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input []byte

		if r.Body != nil {
			defer r.Body.Close()

			body, _ := ioutil.ReadAll(r.Body)

			input = body
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hello world, input was: %s", string(input))))
	})
}
