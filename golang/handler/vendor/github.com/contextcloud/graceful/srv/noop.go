package srv

import "context"

type noop struct {
}

func (n *noop) Start(ctx context.Context) error {
	return nil
}

func (n *noop) Shutdown(ctx context.Context) error {
	return nil
}

func NewNoop() Startable {
	return &noop{}
}
