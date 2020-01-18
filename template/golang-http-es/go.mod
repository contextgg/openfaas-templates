module github.com/contextgg/openfaas-templates/template/golang-http-es

go 1.13

// replace github.com/contextcloud/templates/template/golang-http-es/function => ./handler/function

require (
	github.com/contextgg/go-es v1.6.1
	github.com/contextgg/go-sdk v1.6.5
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/tidwall/pretty v1.0.0 // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
)
