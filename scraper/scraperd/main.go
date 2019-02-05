package main

import (
	"flag"

	"github.com/Emyrk/go-factom-vote/vote/database"

	"github.com/Emyrk/go-factom-vote/scraper"
	log "github.com/sirupsen/logrus"
)

type arrayFlags []string

var version = "v1.0.0"

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	enabledRoutines := arrayFlags{}

	var (
		factomdhost = flag.String("fhost", "localhost", "Factomd host")
		factomdport = flag.Int("fport", 8088, "Factomd port")

		postgreshost = flag.String("phost", "localhost", "Postgres host")
		postgresport = flag.Int("pport", 5432, "Postgres port")
	)

	// For Debugging
	flag.Var(&enabledRoutines, "routine", "Can modify which routines are run")
	flag.Parse()

	config := new(database.SqlConfig)
	if *postgreshost != "localhost" {
		config.SqlConfigType = database.SQL_CON_CUSTOM
		config.User = "postgres"
		config.Pass = "password"
		config.Host = *postgreshost
		config.Port = *postgresport
		config.Schema = database.SCHEMA_PUBLIC
	} else {
		config = nil
	}

	s, err := scraper.NewScraper(*factomdhost, *factomdport, config)
	if err != nil {
		panic(err)
	}

	log.Infof("Running Scraper %s", version)

	if len(enabledRoutines) == 0 {
		enabledRoutines = []string{"catchup"}
	}

	// Does as goroutine if not last
	do := func(f func(), i, l int) {
		if i == l {
			f()
		} else {
			go f()
		}
	}

	// Kinda hacky, but allows me to only run 1 routine if I want.
	for i, r := range enabledRoutines {
		switch r {
		case "catchup":
			do(s.Catchup, i, len(enabledRoutines)-1)
		}
	}
}
