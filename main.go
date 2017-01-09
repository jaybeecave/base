package main

import (
	"html/template"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jaybeecave/render"

	"github.com/johntdyer/slackrus"

	"github.com/jaybeecave/base/datastore"
	"github.com/jaybeecave/base/settings"
	"github.com/jaybeecave/base/viewbucket"
	"github.com/urfave/negroni"
)

func main() {
	// load settings including ENV Variables
	settings := settings.New()

	// - LOGGING
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)
	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
	logHook := &slackrus.SlackrusHook{
		HookURL:        "https://hooks.slack.com/services/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",
		AcceptedLevels: slackrus.LevelThreshold(log.DebugLevel),
		Channel:        "#" + settings.Sitename + "-logs",
		Username:       settings.ServerIs,
	}
	if settings.ServerIsDEV {
		logHook.IconEmoji = ":hamster:"
	} else {
		logHook.IconEmoji = ":dog:"
	}
	log.AddHook(logHook)
	log.Info("App Started. Server Is: " + settings.ServerIs)

	store := datastore.New()
	store.Settings = settings

	n := negroni.Classic() // Includes some default middlewares

	var renderer = render.New(render.Options{
		Layout:     "application",
		Extensions: []string{".html"},
		Funcs:      []template.FuncMap{viewbucket.TemplateFunctions},
		// prevent having to rebuild for every template reload... This is an important setting for development speed
		IsDevelopment:   store.Settings.ServerIsDEV,
		RequirePartials: store.Settings.ServerIsDEV,
		RequireBlocks:   store.Settings.ServerIsDEV,
	})

	n.UseHandler(routes(renderer, store))
	n.Run(settings.ServerPort)
}
