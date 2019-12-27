module handler

go 1.12

replace handler/function => ./function

require (
	github.com/contextgg/go-es v1.5.0
	github.com/contextgg/go-sdk v1.6.2
	github.com/nats-io/nats-server/v2 v2.1.2 // indirect
)
