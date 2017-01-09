package settings

import (
	"os"

	dotenv "github.com/joho/godotenv"
)

// Settings - common settings used around the site. Currently loaded into the datastore object
type Settings struct {
	ServerIsDEV bool
	ServerIsLVE bool
	ServerIs    string
	DSN         string
	Sitename    string
	EncKey      string
	ServerPort  string
}

// New - returns a settings object to be used globally - not as a global variable, but passed around via the datastore.
func New() *Settings {
	err := dotenv.Load()
	if err != nil {
		panic(err)
	}
	s := &Settings{}
	s.ServerIsDEV = (os.Getenv("IS_DEV") == "true")
	s.ServerIsLVE = !s.ServerIsDEV
	if s.ServerIsDEV {
		s.ServerIs = "DEV"
	}
	s.DSN = os.Getenv("DATABASE_URL")
	s.Sitename = os.Getenv("SITE_NAME")
	s.EncKey = os.Getenv("SECURITY_ENCRYPTION_KEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = ":80"
	} else {
		port = ":" + port // append the :
	}
	s.ServerPort = port
	return s
}
