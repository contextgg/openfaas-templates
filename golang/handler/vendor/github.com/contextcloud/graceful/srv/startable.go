package srv

import (
	"context"
	"fmt"
	"net/http"
)

type Startable interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func NewStartable(srvAddr string, h interface{}) (Startable, error) {
	switch start := h.(type) {
	case Startable:
		return start, nil
	case http.Handler:
		handler := WithMetricsRecorder(start)
		return NewStandard(srvAddr, handler), nil
	default:
		return nil, fmt.Errorf("unknown service type: %T", h)
	}
}
