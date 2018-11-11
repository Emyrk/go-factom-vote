package apiserver

import "github.com/Emyrk/go-factom-vote/vote/database"

type GraphQLServer struct {
	SQLDB GraphQLSQLDB
}

func NewGraphQLServer() (*GraphQLServer, error) {
	s := new(GraphQLServer)
	db, err := database.InitLocalDB()
	if err != nil {
		return nil, err
	}

	s.SQLDB.SQLDatabase = db

	return s, nil
}
