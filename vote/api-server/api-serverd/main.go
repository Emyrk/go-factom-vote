package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Emyrk/go-factom-vote/vote/api-server"
	"github.com/Emyrk/go-factom-vote/vote/database"
	"github.com/graphql-go/handler"
)

func main() {

	var (
		postgreshost = flag.String("phost", "localhost", "Postgres host")
		postgresport = flag.Int("pport", 5432, "Postgres port")
	)

	flag.Parse()

	config := new(database.SqlConfig)

	config.SqlConfigType = database.SQL_CON_CUSTOM
	config.User = "postgres"
	config.Pass = "password"
	config.Host = *postgreshost
	config.Port = *postgresport
	config.Schema = database.SCHEMA_PUBLIC
	if *postgreshost != "localhost" {
		config.SqlConfigType = database.SQL_CON_LOCAL
	}

	srv, err := apiserver.NewGraphQLServer(*config)
	if err != nil {
		panic(err)
	}

	schema, err := srv.CreateSchema()
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
