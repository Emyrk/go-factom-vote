package apiserver

import (
	"log"

	"github.com/graphql-go/graphql"
)

func (s *GraphQLServer) CreateSchema() (graphql.Schema, error) {
	// Schema
	fields := graphql.Fields{
		"completed":      s.completedField(),
		"proposal":       s.proposal(),
		"allProposals":   s.allProposals(),
		"eligibleList":   s.eligibleList(),
		"eligibleVoters": s.eligibleListVoters(),
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return schema, err
}

func (s *GraphQLServer) completedField() *graphql.Field {
	return &graphql.Field{
		Type: graphql.String,
		Resolve: func(q graphql.ResolveParams) (interface{}, error) {
			return s.SQLDB.FetchHighestDBInserted(), nil
		},
	}
}

func (s *GraphQLServer) proposal() *graphql.Field {
	return &graphql.Field{
		Type: VoteGraphQLType,
		Args: graphql.FieldConfigArgument{
			"chain": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			chain := params.Args["chain"].(string)
			return s.SQLDB.FetchVote(chain)
		},
	}
}

func (s *GraphQLServer) allProposals() *graphql.Field {
	return &graphql.Field{
		Type: VoteListGraphQLType,
		Args: graphql.FieldConfigArgument{
			"registered": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"active": &graphql.ArgumentConfig{
				Type: graphql.Boolean,
			},
			"offset": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"limit": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(params graphql.ResolveParams) (interface{}, error) {
			reg, _ := params.Args["registered"].(bool)
			act, _ := params.Args["active"].(bool)
			offset, _ := params.Args["offset"].(int)
			limit, _ := params.Args["limit"].(int)

			return s.SQLDB.FetchAllVotes(reg, act, limit, offset)
		},
	}
}

func (s *GraphQLServer) eligibleList() *graphql.Field {
	return &graphql.Field{
		Type: JSON,
		Args: graphql.FieldConfigArgument{
			"chain": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			chainid := p.Args["chain"].(string)
			list, err := s.SQLDB.FetchEligibleList(chainid)
			if err != nil {
				return nil, err
			}
			return list.Admin, nil
		},
	}
}

//func (g *GraphQLServer) EligibleListGraphQLType() *graphql.Object {
//	return graphql.NewObject(graphql.ObjectConfig{
//		Name: "EligbleListAdmin",
//		Fields: graphql.Fields{
//			"admin": &graphql.Field{
//				Type: ELAdminGraphQLType,
//				Args: graphql.FieldConfigArgument{
//					"chain": &graphql.ArgumentConfig{
//						Type: graphql.NewNonNull(graphql.String),
//					},
//				},
//				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
//					chainid := p.Args["chain"].(string)
//					list, err := g.SQLDB.FetchEligibleList(chainid)
//					if err != nil {
//						return nil, err
//					}
//					return list.Admin, nil
//				},
//			},
//			"voterList": g.eligibleListVoters(),
//		}})
//}

func (g *GraphQLServer) eligibleListVoters() *graphql.Field {
	return &graphql.Field{
		Type: ELContainerGraphQLType,
		Name: "VoterList",
		Args: graphql.FieldConfigArgument{
			"chain": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
			"offset": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"limit": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			chainid := p.Args["chain"].(string)
			offset, _ := p.Args["offset"].(int)
			limit, _ := p.Args["limit"].(int)
			return g.SQLDB.FetchEligibleVoters(chainid, limit, offset)
		},
	}
}
