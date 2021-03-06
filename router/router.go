package router

import (
	"context"
	"net/http"

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

func New(renderer *render.Render, store *datastore.Datastore) *CustomRouter {
	customRouter := &CustomRouter{}
	r := bone.New()
	r.CaseSensitive = false
	r.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	r.Handle("/webapp/", http.StripPrefix("/webapp/", http.FileServer(http.Dir("./webapp/js"))))
	r.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	r.Handle("/attachments/", http.StripPrefix("/attachments/", http.FileServer(http.Dir(store.Settings.AttachmentsFolder))))
	customRouter.Router = r
	customRouter.Renderer = renderer
	customRouter.Store = store
	return customRouter
}

// GET - Get handler
func (customRouter *CustomRouter) GET(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.GetFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// POST - Post handler
func (customRouter *CustomRouter) POST(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.PostFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// PST - Post handler with pst for tidier lines
func (customRouter *CustomRouter) PST(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.PostFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// PUT - Put handler
func (customRouter *CustomRouter) PUT(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.PutFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// PATCH - Patch handler
func (customRouter *CustomRouter) PATCH(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.PatchFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// OPTIONS - Options handler
func (customRouter *CustomRouter) OPTIONS(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.OptionsFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// DELETE - Delete handler
func (customRouter *CustomRouter) DELETE(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.DeleteFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

// DEL - Delete handler
func (customRouter *CustomRouter) DEL(route string, routeFunc CustomHandlerFunc, securityType string) *bone.Route {
	//route = strings.ToLower(route)
	return customRouter.Router.DeleteFunc(route, handler(customRouter.Renderer, customRouter.Store, routeFunc, securityType))
}

func handler(renderer *render.Render, store *datastore.Datastore, fn CustomHandlerFunc, authMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		padlock := security.New(req, store)

		// check for a logged in user. We always check this incase we need it
		loggedInUser, _ := padlock.LoggedInUser()
		// if loggedInUser != nil {
		// 	store.ViewGlobals["User"] = loggedInUser // concurrency issues....
		// 	store.ViewGlobals["IsLoggedIn"] = true
		// }

		// req.WithContext()
		// check the route requires auth
		if authMethod == security.NoAuth {
			fn(w, req, renderer, store)
			return
		}

		// if we are at this point then we want a login
		if loggedInUser != nil {
			ctx := context.WithValue(req.Context(), "loggedInUser", loggedInUser)
			fn(w, req.WithContext(ctx), renderer, store)
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
