package function

import (
	"context"
	"log"

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
