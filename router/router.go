package router

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/go-zoo/bone"
	"github.com/jaybeecave/base/datastore"
	"github.com/jaybeecave/base/security"
	"github.com/jaybeecave/render"
)

// CustomRouter wraps gorilla mux with database, redis and renderer
type CustomRouter struct {
	// Router *mux.Router
	Router   *bone.Mux
	Renderer *render.Render
	Store    *datastore.Datastore
}

// New - Create a new custom router instance
func NewWithConsole(renderer *render.Render, store *datastore.Datastore) *CustomRouter {
	customRouter := New(renderer, store)
	if store.Settings.ServerIsDEV {
	}
	return customRouter
}

func New(renderer *render.Render, store *datastore.Datastore) *CustomRouter {
	customRouter := &CustomRouter{}
	r := bone.New()
	r.CaseSensitive = false
	customRouter.Router = r
	customRouter.Renderer = renderer
	customRouter.Store = store
	return customRouter
}

// func (customRouter *CustomRouter) Route(route string, routeFunc CustomHandlerFunc, securityType string) *mux.Route {
// 	return customRouter.Router.HandleFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
// }

// GET - Get handler
func (customRouter *CustomRouter) GET(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.GetFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
	// return customRouter.Router.HandleFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// POST - Post handler
func (customRouter *CustomRouter) POST(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.PostFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// PUT - Put handler
func (customRouter *CustomRouter) PUT(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.PutFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// PATCH - Patch handler
func (customRouter *CustomRouter) PATCH(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.PatchFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// OPTIONS - Options handler
func (customRouter *CustomRouter) OPTIONS(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.OptionsFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// DELETE - Delete handler
func (customRouter *CustomRouter) DELETE(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	return customRouter.Router.DeleteFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

func handler(renderer *render.Render, store *datastore.Datastore, fn CustomHandlerFunc, authMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		padlock := security.New(req, store)

		// check the route requires auth
		if authMethod == security.NoAuth {
			fn(w, req, renderer, store)
			return
		}

		// check the user is authenticated
		isLoggedIn, err := padlock.CheckLogin()
		if err != nil {
			if isLoggedIn {
				panic("logged in with an error in security. How the hell did that happen! #REF-35353")
			}
			log.Error("login failed #REF-73644: ", err)
		}

		if isLoggedIn {
			fn(w, req, renderer, store)
			return
		}

		// if we have reached this point then the user doesn't have access
		if authMethod == security.Disallow {
			renderer.JSON(w, http.StatusForbidden, "Not logged in")
		} else if authMethod == security.Redirect {
			http.Redirect(w, req, "/Login", http.StatusFound)
		}
	}
}

type CustomHandlerFunc func(http.ResponseWriter, *http.Request, *render.Render, *datastore.Datastore)
