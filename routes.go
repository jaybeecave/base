package main

import (
	"github.com/go-zoo/bone"
	store "github.com/jaybeecave/base/datastore"
	"github.com/jaybeecave/base/router"
	"github.com/jaybeecave/render"
)

func routes(renderer *render.Render, datastore *store.Datastore) *bone.Mux {
	r := router.New(renderer, datastore)

	// // Login routes
	// r.GET("/login", login, security.NoAuth)
	// r.POST("/login-submit", loginSubmit, security.NoAuth)
	// r.GET("/logged-in", loggedIn, security.Disallow)

	// Scaffold routes

	return r.Router
}
