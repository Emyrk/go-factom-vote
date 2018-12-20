package apiserver

import (
	"fmt"

	"github.com/Emyrk/go-factom-vote/vote/database"
	"github.com/FactomProject/factom"
)

type GraphQLServer struct {
	SQLDB GraphQLSQLDB
}

func NewGraphQLServer(sqlConfig database.SqlConfig, factomHost string, factomPort int) (*GraphQLServer, error) {
	s := new(GraphQLServer)
	db, err := database.InitDb(sqlConfig)
	if err != nil {
		return nil, err
	}

	s.SQLDB.SQLDatabase = db

	factom.SetFactomdServer(fmt.Sprintf("%s:%d", factomHost, factomPort))

	return s, nil
}
