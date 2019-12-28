package function

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/contextgg/go-es/builder"
	"github.com/contextgg/go-es/es"
)

// HelloCommand our hello command
type HelloCommand struct {
	es.BaseCommand
	Text string `json:"text"`
}

type HelloCommandHandler struct {
}

func (h *HelloCommandHandler) HandleCommand(ctx context.Context, cmd es.Command) error {
	hello := cmd.(*HelloCommand)
	log.Printf("Hello, %s", hello.Text)
	return nil
}

func Setup(build builder.ClientBuilder) {
	build.WireCommandHandler(&HelloCommandHandler{}, builder.Command(&HelloCommand{}))
}

// NewHandler creates a new http handler
func NewHandler(client *builder.Client) http.Handler {
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
