package apiserver

import "github.com/Emyrk/go-factom-vote/vote/database"

type GraphQLServer struct {
	SQLDB GraphQLSQLDB
}

func NewGraphQLServer(sqlConfig database.SqlConfig) (*GraphQLServer, error) {
	s := new(GraphQLServer)
	db, err := database.InitDb(sqlConfig)
	if err != nil {
		return nil, err
	}

	s.SQLDB.SQLDatabase = db

	return s, nil
}
