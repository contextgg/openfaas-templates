package faas

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/contextgg/go-sdk/autha"
	"github.com/contextgg/go-sdk/httpbuilder"
)

// PersistCommand struct
type PersistCommand struct {
	*autha.PersistUser

	AggregateID string `json:"aggregate_id"`
}

// ConnectCommand struct
type ConnectCommand struct {
	*autha.ConnectUser

	AggregateID string `json:"aggregate_id"`
}

type provider struct {
	functionName string
	namespace    uuid.UUID

	debug bool
}

func (p *provider) SetDebug() {
	p.debug = true
}

func (p *provider) CalculateAggregateID(connection, id string) string {
	ns := uuid.NewSHA1(p.namespace, []byte(connection))
	uid := uuid.NewSHA1(ns, []byte(id))
	return uid.String()
}

// connection string, id *autha.Identity, token autha.Token
func (p *provider) Persist(ctx context.Context, aggregateID string, m *autha.PersistUser) error {
	raw := PersistCommand{
		m,
		aggregateID,
	}

	if p.debug {
		log.Printf("Payload: %#v", raw)
	}

	var errorString string
	status, err := httpbuilder.NewFaaS().
		SetFunction(p.functionName).
		AppendPath("persistuser").
		SetMethod(http.MethodPost).
		SetBody(&raw).
		SetErrorString(&errorString).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("Error with request %w %s", err, errorString)
	}
	if status < 200 || status > 399 {
		return fmt.Errorf("Error with status code %d %s", status, errorString)
	}

	return nil
}

// connection string, id *autha.Identity, token autha.Token
func (p *provider) Connect(ctx context.Context, aggregateID string, m *autha.ConnectUser) error {
	raw := ConnectCommand{
		m,
		aggregateID,
	}
	if p.debug {
		log.Printf("Payload: %#v", raw)
	}

	var errorString string
	status, err := httpbuilder.NewFaaS().
		SetFunction(p.functionName).
		AppendPath("connectuser").
		SetMethod(http.MethodPost).
		SetBody(&raw).
		SetErrorString(&errorString).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("Error with request %w %s", err, errorString)
	}
	if status < 200 || status > 399 {
		return fmt.Errorf("Error with status code %d %s", status, errorString)
	}

	return nil
}

// NewService creates a new user provider
func NewService(functionName, authDNS string) autha.UserService {
	base := uuid.NewSHA1(uuid.NameSpaceURL, []byte(authDNS))

	return &provider{
		functionName: functionName,
		namespace:    base,
	}
}
