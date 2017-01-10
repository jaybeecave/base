package datastore

import (
	"database/sql"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/jaybeecave/base/settings"
	dat "gopkg.in/mgutz/dat.v1"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
	redis "gopkg.in/redis.v5"
)

type Datastore struct {
	DB          *runner.DB
	Cache       *redis.Client
	Settings    *settings.Settings
	ViewGlobals map[string]interface{}
}

var viewGlobals = map[string]interface{}{
	"Date":      time.Now(),
	"Copyright": time.Now().Year(),
}

// New - returns a new datastore which contains redis, database, view globals and settings.
func New() *Datastore {
	store := &Datastore{}
	// THIS MUST BE FIRST it loads ENV variables

	store.DB = getDBConnection()
	store.Cache = getCacheConnection()
	store.ViewGlobals = viewGlobals
	return store
}

func getDBConnection() *runner.DB {
	//get url from ENV in the following format postgres://user:pass@192.168.8.8:5432/spaceio")
	dbURL := os.Getenv("DATABASE_URL")
	u, err := url.Parse(dbURL)
	if err != nil {
		log.Error(err)
	}

	username := u.User.Username()
	pass, isPassSet := u.User.Password()
	if !isPassSet {
		log.Error("no database password")
	}
	host, port, _ := net.SplitHostPort(u.Host)
	dbName := strings.Replace(u.Path, "/", "", 1)

	db, _ := sql.Open("postgres", "dbname="+dbName+" user="+username+" password="+pass+" host="+host+" port="+port+" sslmode=disable")
	err = db.Ping()
	if err != nil {
		log.Error(err)
	}
	log.Info("database running")
	// ensures the database can be pinged with an exponential backoff (15 min)
	runner.MustPing(db)

	// set to reasonable values for production
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)

	// set this to enable interpolation
	dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	// Log any query over 10ms as warnings. (optional)
	runner.LogQueriesThreshold = 10 * time.Millisecond

	// db connection
	return runner.NewDB(db, "postgres")
}

func getCacheConnection() *redis.Client {

	opts := &redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	url := os.Getenv("REDIS_URL")
	if url != "" {
		newOpts, err := redis.ParseURL(url)
		if err == nil {
			opts = newOpts
		} else {
			log.Error(err)
			return nil
		}
	}

	client := redis.NewClient(opts)
	pong, err := client.Ping().Result()
	if err != nil {
		log.Error(err)
		return nil
	}

	log.Info("cache running", pong)
	return client
}
