package main

import (
	"log"
	"net/http"

	"fmt"

	"github.com/Emyrk/go-factom-vote/vote/database"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func main() {

	db, err := database.InitLocalDB()
	if err != nil {
		panic(err)
	}

	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				fmt.Println(p.Args)
				//return db.IsEligibleListExist(p.Args)
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	var _ = db

	//// Query
	//query := `
	//	{
	//		hello
	//	}
	//`
	//params := graphql.Params{Schema: schema, RequestString: query}
	//r := graphql.Do(params)
	//if len(r.Errors) > 0 {
	//	log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	//}
	//rJSON, _ := json.Marshal(r)
	//fmt.Printf("%s \n", rJSON) // {“data”:{“hello”:”world”}}
	//
	//
	//schema, _ := graphql.NewSchema(...)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
