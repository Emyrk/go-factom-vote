package apiserver

import (
	"encoding/json"

	"fmt"

	"database/sql"

	"github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

type ProposalEntry struct {
	VoterId   string  `json:"voterId"`
	Weight    float64 `json:"weight"`
	EntryHash string  `json:"entryHash"`

	// These can be null
	Commit sql.NullString `json:"commit"`
	Reveal sql.NullString `json:"reveal"`
}

func NewProposalEntry() *ProposalEntry {
	p := new(ProposalEntry)

	return p
}

var FactomdProperties = graphql.NewObject(graphql.ObjectConfig{
	Name:        "FactomdProperties",
	Description: "Factomd Version",
	Fields: graphql.Fields{
		"factomdVersion": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				lst, ok := p.Source.([]string)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}
				return lst[0], nil
			},
		},
		"factomdVersionError": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				lst, ok := p.Source.([]string)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}
				return lst[1], nil
			},
		},
		"factomdAPIVersion": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				lst, ok := p.Source.([]string)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}
				return lst[2], nil
			},
		},
		"factomdAPIVersionError": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				lst, ok := p.Source.([]string)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}
				return lst[3], nil
			},
		},
	}})

var ProposalEntryGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ProposalEntry",
	Description: "A commit or reveal in a proposal",
	Fields: graphql.Fields{
		"voterId": &graphql.Field{
			Type: graphql.String,
		},
		"entryHash": &graphql.Field{
			Type:        graphql.String,
			Description: "EntryHash of the entry that added the voter to the eligible list",
		},
		"weight": &graphql.Field{
			Type:        graphql.Float,
			Description: "Voter's voting weight",
		},
		"commit": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ent, ok := p.Source.(ProposalEntry)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}

				str := ent.Commit
				if str.Valid {
					return str.String, nil
				}
				return nil, nil
			},
		},
		"reveal": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ent, ok := p.Source.(ProposalEntry)
				if !ok {
					return nil, fmt.Errorf("Incorrect type supplied")
				}

				str := ent.Reveal
				if str.Valid {
					return str.String, nil
				}
				return nil, nil
			},
		},
	}})

type ListInfo struct {
	TotalCount int `json:"totalCount"`
	Offset     int `json:"offset""`
	Limit      int `json:"limit"`
}

type VoteList struct {
	Info  ListInfo `json:"listInfo"`
	Votes []Vote   `json:"voteList"`
}

type VoteResultList struct {
	Info  ListInfo           `json:"listInfo"`
	Votes []common.VoteStats `json:"resultList"`
}

var VoteListGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "VoteList",
	Description: "A list of votes",
	Fields: graphql.Fields{
		"listInfo": &graphql.Field{
			Type: JSON,
		},
		"voteList": &graphql.Field{
			Type: graphql.NewList(VoteGraphQLType),
		},
	}})

type Vote struct {
	Chainid    string         `json:"voteChainId"`
	Admin      VoteAdmin      `json:"admin"`
	Definition VoteDefinition `json:"vote"`
	Results    VoteResult     `json:"result"`

	Proposal VoteDetails `json:"proposal"` // Title, description, etc
}

var VoteGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Vote",
	Description: "The full vote data structure.",
	Fields: graphql.Fields{
		"voteChainId": &graphql.Field{
			Type: graphql.String,
		},
		"vote": &graphql.Field{
			Type: VoteDefinitionGraphQLType,
		},
		"admin": &graphql.Field{
			Type: VoteAdminGraphQLType,
		},
		"proposal": &graphql.Field{
			Type: VAVoteInfoGraphQLType,
		},
	}})

type VoteAdmin struct {
	// Related to Security
	VoteInitator     string `json:"voteInitiator"`
	SigningKey       string `json:"signingKey"`
	Signature        string `json:"signature"`
	AdminEntryHash   string `json:"adminEntryHash"`
	AdminBlockHeight int    `json:"blockHeight"`
	Registered       bool   `json:"registered"`
	Complete         bool   `json:"complete"`
	Status           string `json:"status"`
	ProtocolVersion  int    `json:"protocolVersion"`

	// Other
	//VoteInfo struct {
	//	Title       string            `json:"title"`
	//	Text        string            `json:"text"`
	//	ExternalRef ExternalReference `json:"externalRef"`
	//} `json:"voteInfo"` // Title, description, etc
}

type VoteDetails struct {
	Title       string            `json:"title"`
	Text        string            `json:"text"`
	ExternalRef ExternalReference `json:"externalRef"`
}

type ExternalReference struct {
	Href string `json:"href"`
	Hash struct {
		Value string `json:"value"`
		Algo  string `json:"algo"`
	} `json:"hash"`
}

type VoteDefinition struct {
	PhasesBlockHeights struct {
		CommitStart int `json:"commitStart"`
		CommitStop  int `json:"commitEnd"`
		RevealStart int `json:"revealStart"`
		RevealStop  int `json:"revealEnd"`
	} `json:"phasesBlockHeights"`
	VoteType           int          `json:"type"`
	Config             GQVoteConfig `json:"config"`
	EligibleVoterChain string       `json:"eligibleVotersChainId"`

	//VoteInfo VoteDetails `json:"proposal"` // Title, description, etc
}

// Uses strings instead of full objects
type GQVoteConfig struct {
	Options               []string `json:"options"`
	MinOptions            int      `json:"minOptions"`
	MaxOptions            int      `json:"maxOptions"`
	AcceptanceCriteria    string   `json:"acceptanceCriteria"`
	WinnerCriteria        string   `json:"winnerCriteria"`
	AllowAbstention       bool     `json:"allowAbstention"`
	ComputeResultsAgainst string   `json:"computeResultsAgainst"`
}

var VoteAdminGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "VoteAdmin",
	Description: "",
	Fields: graphql.Fields{
		"voteInitiator": &graphql.Field{
			Type: graphql.String,
		},
		"signingKey": &graphql.Field{
			Type: graphql.String,
		},
		"signature": &graphql.Field{
			Type: graphql.String,
		},
		"adminEntryHash": &graphql.Field{
			Type: graphql.String,
		},
		"blockHeight": &graphql.Field{
			Type: graphql.Int,
		},
		"registered": &graphql.Field{
			Type: graphql.Boolean,
		},
		"complete": &graphql.Field{
			Type: graphql.Boolean,
		},
		"protocolVersion": &graphql.Field{
			Type: graphql.Int,
		},
		//"voteInfo": &graphql.Field{
		//	Type: VAVoteInfoGraphQLType,
		//},
	}})

var VAVoteInfoGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "VoteInfo",
	Description: "",
	Fields: graphql.Fields{
		"title": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				vd, ok := p.Source.(VoteDetails)
				if !ok {
					return nil, fmt.Errorf("Bad type")
				}
				if vd.Title == "" {
					return nil, nil
				}

				return vd.Title, nil
			},
		},
		"text": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				vd, ok := p.Source.(VoteDetails)
				if !ok {
					return nil, fmt.Errorf("Bad type")
				}
				if vd.Text == "" {
					return nil, nil
				}

				return vd.Text, nil
			},
		},
		"externalRef": &graphql.Field{
			Type: JSON,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				vd, ok := p.Source.(VoteDetails)
				if !ok {
					return nil, fmt.Errorf("Bad type")
				}
				ref := vd.ExternalRef
				if ref.Href == "" &&
					(ref.Hash.Value == "0000000000000000000000000000000000000000000000000000000000000000" || ref.Hash.Value == "") &&
					ref.Hash.Algo == "" {
					return nil, nil
				}

				return ref, nil
			},
		},
	}})

var VoteDefinitionGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "VoteDefinition",
	Description: "The vote definition",
	Fields: graphql.Fields{
		"phasesBlockHeights": &graphql.Field{
			Type: PhasesBlockGraphQLType,
		},
		"type": &graphql.Field{
			Type: graphql.Int,
		},
		"config": &graphql.Field{
			Type: VDConfigGraphQLType,
		},
		"eligibleVotersChainId": &graphql.Field{
			Type: graphql.String,
		},
	}})

var PhasesBlockGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "PhasesBlocks",
	Description: "The block to distinguish the block heigths for the various phases of voting",
	Fields: graphql.Fields{
		"commitStart": &graphql.Field{
			Type: graphql.Int,
		},
		"commitEnd": &graphql.Field{
			Type: graphql.Int,
		},
		"revealStart": &graphql.Field{
			Type: graphql.Int,
		},
		"revealEnd": &graphql.Field{
			Type: graphql.Int,
		},
	}})

//var VDPhasesBlockHeightsGraphQLType = graphql.NewObject(graphql.ObjectConfig{
//	Name:        "PhasesBlockHeights",
//	Description: "",
//	Fields: graphql.Fields{
//		"commitStart": &graphql.Field{
//			Type: graphql.Int,
//		},
//		"commitStop": &graphql.Field{
//			Type: graphql.Int,
//		},
//		"revealStart": &graphql.Field{
//			Type: graphql.Int,
//		},
//		"revealStop": &graphql.Field{
//			Type: graphql.Int,
//		},
//	}})

var VDConfigGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "VoteDefinitionConfig",
	Description: "",
	Fields: graphql.Fields{
		"options": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"minOptions": &graphql.Field{
			Type: graphql.Int,
		},
		"maxOptions": &graphql.Field{
			Type: graphql.Int,
		},
		"acceptanceCriteria": &graphql.Field{
			Type: JSON,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				s := p.Source.(GQVoteConfig)
				crit := common.AcceptCriteriaStruct{}
				err := json.Unmarshal([]byte(s.AcceptanceCriteria), &crit)
				if err != nil {
					return nil, err
				}

				if crit.MinTurnout.Weighted+crit.MinTurnout.Unweighted == 0 {
					return nil, nil
				}

				return crit, nil
			},
		},
		"winnerCriteria": &graphql.Field{
			Type: JSON,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				s := p.Source.(GQVoteConfig)
				crit := common.WinnerCriteriaStruct{}
				err := json.Unmarshal([]byte(s.WinnerCriteria), &crit)
				if err != nil {
					return nil, err
				}
				return crit, nil
			},
		},
		"allowAbstention": &graphql.Field{
			Type: graphql.Boolean,
		},
		"computeResultsAgainst": &graphql.Field{
			Type: graphql.String,
		},
	}})

/*
 *
 * Vote Result
 *
 */

type VoteResult struct {
}

/*
 *
 * Commit/Reveal
 *
 */

type VoteCommitContainer struct {
	Commits []VoteCommit `json:"commits"`
	Info    ListInfo     `json:"listInfo"`
}

var CommitListGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CommitList",
	Description: "A list of commits",
	Fields: graphql.Fields{
		"listInfo": &graphql.Field{
			Type: JSON,
		},
		"commits": &graphql.Field{
			Type: graphql.NewList(VoteCommitGraphQLType),
		},
	}})

type VoteCommit struct {
	VoterID     string `json:"voterId"`
	VoteChain   string `json:"voteChain"`
	SigningKey  string `json:"signingKey"`
	Signature   string `json:"signature"`
	Commitment  string `json:"commitment"`
	EntryHash   string `json:"entryhash"`
	BlockHeight int    `json:"blockHeight"`
}

var VoteCommitGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "VoteCommit",
	Fields: graphql.Fields{
		"voterId": &graphql.Field{
			Type: graphql.String,
		},
		"voteChain": &graphql.Field{
			Type: graphql.String,
		},
		"signingKey": &graphql.Field{
			Type: graphql.String,
		},
		"signature": &graphql.Field{
			Type: graphql.String,
		},
		"commitment": &graphql.Field{
			Type: graphql.String,
		},
		"entryhash": &graphql.Field{
			Type: graphql.String,
		},
		"blockHeight": &graphql.Field{
			Type: graphql.Int,
		},
	}})

type VoteRevealContainer struct {
	Reveals []VoteReveal `json:"reveals"`
	Info    ListInfo     `json:"listInfo"`
}

var RevealListGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "RevealList",
	Description: "A list of reveals",
	Fields: graphql.Fields{
		"listInfo": &graphql.Field{
			Type: JSON,
		},
		"reveals": &graphql.Field{
			Type: graphql.NewList(VoteRevealGraphQLType),
		},
	}})

type VoteReveal struct {
	VoterID     string   `json:"voterId"`
	VoteChain   string   `json:"voteChain"`
	Vote        []string `json:"vote"`
	Secret      string   `json:"secret"`
	HmacAlgo    string   `json:"hmacAlgo"`
	EntryHash   string   `json:"entryhash"`
	BlockHeight int      `json:"blockHeight"`
}

var VoteRevealGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "VoteReveal",
	Fields: graphql.Fields{
		"voterId": &graphql.Field{
			Type: graphql.String,
		},
		"voteChain": &graphql.Field{
			Type: graphql.String,
		},
		"vote": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
		"secret": &graphql.Field{
			Type: graphql.String,
		},
		"hmacAlgo": &graphql.Field{
			Type: graphql.String,
		},
		"entryhash": &graphql.Field{
			Type: graphql.String,
		},
		"blockHeight": &graphql.Field{
			Type: graphql.Int,
		},
	}})

/*
 *
 * Eligible List
 *
 */

type EligibleList struct {
	Admin struct {
		ChainID    string `json:"chainId"`
		Initiator  string `json:"initiator"`
		Nonce      string `json:"nonce"`
		SigningKey string `json:"signingKey"`
		Signature  string `json:"signature"`
	} `json:"admin"`
}

type EligibleVoterContainer struct {
	EligibleVoters []EligibleVoter `json:"voters"`
	Info           ListInfo        `json:"listInfo"`
}

type EligibleVoter struct {
	// Given by Entry
	VoterID    string  `json:"voterId"`
	VoteWeight float64 `json:"weight"`

	// Given by entry context
	BlockHeight  int    `json:"blockHeight"`
	EligibleList string `json:"eligibleList"`
	EntryHash    string `json:"entryHash"`

	// Given by factom-walletd
	SigningKeys []string `json:"keys"`
}

var ELContainerGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "EligibleList",
	Fields: graphql.Fields{
		"listInfo": &graphql.Field{
			Type: JSON,
		},
		"voters": &graphql.Field{
			Description: "TODO: Should allow this to be broken up",
			Type:        graphql.NewList(ELVoter),
		},
	}})

var ELVoter = graphql.NewObject(graphql.ObjectConfig{
	Name: "EligibleVoter",
	Fields: graphql.Fields{
		"voterId": &graphql.Field{
			Type: graphql.String,
		},
		"weight": &graphql.Field{
			Type: graphql.Float,
		},
		"blockHeight": &graphql.Field{
			Type: graphql.Int,
		},
		"entryHash": &graphql.Field{
			Type: graphql.String,
		},
		"keys": &graphql.Field{
			Type: graphql.NewList(graphql.String),
		},
	}})

var ELAdminGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "EligibleListAdmin",
	Fields: graphql.Fields{
		"chainId": &graphql.Field{
			Type: graphql.String,
		},
		"initiator": &graphql.Field{
			Type: graphql.String,
		},
		"nonce": &graphql.Field{
			Type: graphql.String,
		},
		"signingKey": &graphql.Field{
			Type: graphql.String,
		},
		"signature": &graphql.Field{
			Type: graphql.String,
		},
	}})

//

// Results (Use common one as base)

var VoteResultsListGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "VoteResultsList",
	Fields: graphql.Fields{
		"listInfo": &graphql.Field{
			Type: JSON,
		},
		"resultList": &graphql.Field{
			Type: graphql.NewList(VoteResultsGraphQLType),
		},
	}})

var VoteResultsGraphQLType = graphql.NewObject(graphql.ObjectConfig{
	Name: "VoteResults",
	Fields: graphql.Fields{
		"chainId": &graphql.Field{
			Type: graphql.String,
		},
		"valid": &graphql.Field{
			Type: graphql.Boolean,
		},
		"invalidReason": &graphql.Field{
			Type: graphql.String,
		},
		"total": &graphql.Field{
			Type: JSON,
		},
		"voted": &graphql.Field{
			Type: JSON,
		},
		"abstain": &graphql.Field{
			Type: JSON,
		},
		"options": &graphql.Field{
			Type: JSON,
		},
		"turnout": &graphql.Field{
			Type: JSON,
		},
		"support": &graphql.Field{
			Type: JSON,
		},
		"weightedWinners": &graphql.Field{
			Type: JSON,
		},
	}})

// JSON json type
var JSON = graphql.NewScalar(
	graphql.ScalarConfig{
		Name:        "JSON",
		Description: "The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf)",
		Serialize: func(value interface{}) interface{} {
			return value
		},
		ParseValue: func(value interface{}) interface{} {
			return value
		},
		ParseLiteral: parseLiteral,
	},
)

func parseLiteral(astValue ast.Value) interface{} {
	kind := astValue.GetKind()

	switch kind {
	case kinds.StringValue:
		return astValue.GetValue()
	case kinds.BooleanValue:
		return astValue.GetValue()
	case kinds.IntValue:
		return astValue.GetValue()
	case kinds.FloatValue:
		return astValue.GetValue()
	case kinds.ObjectValue:
		obj := make(map[string]interface{})
		for _, v := range astValue.GetValue().([]*ast.ObjectField) {
			obj[v.Name.Value] = parseLiteral(v.Value)
		}
		return obj
	case kinds.ListValue:
		list := make([]interface{}, 0)
		for _, v := range astValue.GetValue().([]ast.Value) {
			list = append(list, parseLiteral(v))
		}
		return list
	default:
		return nil
	}
}
