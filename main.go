package main

import (
	"html/template"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/unrolled/render"

	"github.com/johntdyer/slackrus"

	"github.com/jaybeecave/base/datastore"
	"github.com/jaybeecave/base/viewbucket"
	"github.com/urfave/negroni"
)

// Start - main starting point of the application
func main() {
	store := datastore.New()

	// - SERVER IS
	serverIs := "DEV"
	emoji := ":hamster:"

	if os.Getenv("IS_DEV") == "" {
		serverIs = "LVE"
		emoji = ":dog:"
	}

	// - LOGGING
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	log.AddHook(&slackrus.SlackrusHook{
		HookURL:        "https://hooks.slack.com/services/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",
		AcceptedLevels: slackrus.LevelThreshold(log.DebugLevel),
		Channel:        "#base-logs",
		IconEmoji:      emoji,
		Username:       serverIs,
	})

	log.Info("App Started")

	n := negroni.Classic() // Includes some default middlewares

	var renderer = render.New(render.Options{
		Layout:        "application",
		Extensions:    []string{".html"},
		Funcs:         []template.FuncMap{viewbucket.TemplateFunctions},
		IsDevelopment: store.Settings.ServerIsDEV,
	})

	n.UseHandler(routes(renderer, store))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	n.Run(":" + port)
}
