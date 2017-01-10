package main

// func login(w http.ResponseWriter, req *http.Request, render *render.Render, store *store.Datastore) {
// 	padlock := security.New(req, store)
// 	isLoggedIn, _ := padlock.CheckLogin()
// 	if isLoggedIn {
// 		flash.Message(w, "InfoMessage", "Hey")
// 		http.Redirect(w, req, "/logged-in", http.StatusFound)
// 	}
// 	view := viewbucket.New(w, req, render, store)
// 	view.Render(http.StatusOK, "login")
// }

// func loginSubmit(w http.ResponseWriter, req *http.Request, render *render.Render, store *store.Datastore) {
// 	helper := models.NewAdministratorHelper(store)
// 	administrator, err := helper.NewFromRequest(req)
// 	if err != nil {
// 		http.Redirect(w, req, "/login", http.StatusFound)
// 		return
// 	}

// 	log.Info(administrator)

// 	padlock := security.New(req, store)
// 	cookie, err := padlock.LoginReturningCookie(administrator.Email, administrator.Password, "administrators")
// 	if err != nil {
// 		http.Redirect(w, req, "/login", http.StatusFound)
// 		return
// 	}
// 	http.SetCookie(w, cookie)
// 	http.Redirect(w, req, "/logged-in", http.StatusFound)
// }

// func loggedIn(w http.ResponseWriter, req *http.Request, render *render.Render, store *store.Datastore) {
// 	view := viewbucket.New(w, req, render, store)
// 	view.Render(http.StatusOK, "logged-in")
// }
