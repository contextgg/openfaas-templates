package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/contextgg/go-sdk/autha"
	"github.com/contextgg/go-sdk/autha/faas"
	"github.com/contextgg/go-sdk/autha/stores"

	_ "github.com/contextgg/go-sdk/autha/providers/battlenet"
	_ "github.com/contextgg/go-sdk/autha/providers/discord"
	_ "github.com/contextgg/go-sdk/autha/providers/smashgg"
	_ "github.com/contextgg/go-sdk/autha/providers/steam"
	_ "github.com/contextgg/go-sdk/autha/providers/twitch"
	_ "github.com/contextgg/go-sdk/autha/providers/twitter"

	"handler/function"
)

func main() {
	cfg := &Config{}
	if err := cfg.Load(); err != nil {
		log.Fatal(err)
		return
	}

	provider := function.NewProvider(cfg.CallbackURL)
	if provider == nil {
		log.Fatal("We require a valid AuthProvider")
		return
	}

	sessionStore, _ := stores.NewSessionStore(cfg.SessionSecure, []byte(cfg.SessionSecret))
	userStore, _ := stores.NewUserStore(cfg.SessionSecure, []byte(cfg.SessionSecret))
	userService := faas.NewService(cfg.UserFunctionName, cfg.DNS)
	auth := autha.NewConfig(cfg.Connection, cfg.LoginURL, cfg.ErrorURL, sessionStore, userStore, provider, userService)

	handler := function.NewHandler(auth)

	s := &http.Server{
		Handler:        handler,
		Addr:           fmt.Sprintf(":%d", 8082),
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}

	log.Fatal(s.ListenAndServe())
}
