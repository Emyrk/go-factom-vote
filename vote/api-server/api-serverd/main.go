package main

import (
	"log"
	"net/http"

	"github.com/Emyrk/go-factom-vote/vote/api-server"
	"github.com/graphql-go/handler"
)

func main() {

	srv, err := apiserver.NewGraphQLServer()
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
