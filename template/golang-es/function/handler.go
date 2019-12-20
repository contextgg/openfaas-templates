package function

import (
	"context"
	"log"

	"github.com/contextgg/go-es/builder"
	"github.com/contextgg/go-es/contrib/basicauth"
	"github.com/contextgg/go-es/es"
	"github.com/contextgg/go-sdk/secrets"
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
	creds := secrets.LoadBasicAuth("auth")

	middleware := []es.CommandHandlerMiddleware{
		basicauth.NewMiddleware(creds),
	}

	build.WireCommandHandler(&HelloCommandHandler{}, builder.Command(&HelloCommand{}, middleware...))
}
