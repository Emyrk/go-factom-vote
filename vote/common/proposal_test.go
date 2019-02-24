package common_test

import (
	"encoding/json"
	"testing"

	"encoding/hex"

	. "github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

//func TestCommitJson(t *testing.T) {
//	j := `
//	{
//		“commitment”: "random"
//	}
//	`
//
//	c := new(VoteCommit)
//	err := json.Unmarshal([]byte(j), c.Content)
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestRevealJson(t *testing.T) {
	j := `
		{
			"vote": ["a", "b"],
			"secret": "5d918148b33f12cb43ce1c5b0cf28a13",
			"hmacAlgo": "sha256"
		}
		`
	r := new(VoteReveal)
	err := json.Unmarshal([]byte(j), r)
	if err != nil {
		t.Error(err)
	}
}

func TestProposalJSONMarshal(t *testing.T) {
	j := `
	{
		"proposal": {
			"title": "The quick fox jumps over the lazy dog.",
			"text": "Testing, Testing, 1, 2, 3",
			"externalRef": {
				"href": "https://google.com",
				"hash": {
					"value": "5d918148b33f12cb43ce1c5b0cf28a1392fa3559f9fc140a657e147b1272c4a9",
					"algo": "sha256"
				}
			}	
		},
		"vote": {
			"phasesBlockHeights": {
				"commitStart": 0,
				"commitEnd": 50,
				"revealStart": 50,
				"revealEnd": 100
			},
			"eligibleVotersChainId": "e90084108491265b700342e83646e6116fdc3a519d07c038bd200ff5fade5570",
			"type": 0,
			"config": {
				"options": [],
				"allowAbstention": false,
				"computeResultsAgainst": "PARTICIPANTS_ONLY",
				"minOptions": 0,
				"maxOptions": 0,
				"acceptanceCriteria": {
					"minSupport": {
						"optionA": {
							"weighted": 1,
							"unweighted": 10
						}
					}
				}
			}
		}
	}`

	p := new(ProposalEntry)
	err := json.Unmarshal([]byte(j), p)
	if err != nil {
		t.Error(err)
	}
}

func TestNewValidVoteCommit(t *testing.T) {
	e := entryBlock.NewEntry()
	chain, _ := primitives.HexToHash("6ac365f648477399de0513e7754902a826c6f6527bfcfaafa1379038e2bae4d3")

	// Ehash on testnet = 50b20e1e0c2be9e46e0eb4c834c213ef4ae2ae24513c1a7347f4871863e5fd1c
	e.ChainID = chain
	e.ExtIDs = make([]primitives.ByteSlice, 4)
	e.ExtIDs[0].Bytes = []byte("factom-vote-commit")
	e.ExtIDs[1].Bytes, _ = hex.DecodeString("2d98021e3cf71580102224b2fcb4c5c60595e8fdf6fd1b97c6ef63e9fb3ed635")
	e.ExtIDs[2].Bytes, _ = hex.DecodeString("c103756200d0c1223c0ee9911196bf06de6ee570b0f45897e2ef39f9abf39d24")
	e.ExtIDs[3].Bytes, _ = hex.DecodeString("a3710899ab19e7142354c4c47712c0f3497d37fb7b077c72f166ec7239560f98f7383ccbc68a8d064e31ef8db572c50f021085be9270ebfffed4983f879cf607")

	e.Content.Bytes = []byte(`{"commitment":"2ddbebd99f1ae15fea4066a9a81087ccb4ce233e26ea6f3599c1eb981d780435b7669e8be41988f15eba021c7a94cd02a17372dd90cd491f52706b5731e61d9e"}`)

	_, err := NewVoteCommitFromEntry(e, 49595)
	if err != nil {
		t.Error(err)
	}
}

// Requires running factomd
//func TestNewValidVoteProposal(t *testing.T) {
//	e := entryBlock.NewEntry()
//	chain, _ := primitives.HexToHash("6ac365f648477399de0513e7754902a826c6f6527bfcfaafa1379038e2bae4d3")
//
//	// Ehash on testnet = b0d50e804e90cd2a7c2e775e2dcf9f4adeeb806e7e84179642ea65cd0c8d3e6b
//	e.ChainID = chain
//	e.ExtIDs = make([]primitives.ByteSlice, 5)
//	e.ExtIDs[0].Bytes = []byte("factom-vote")
//	e.ExtIDs[1].Bytes, _ = hex.DecodeString("fffd")
//	e.ExtIDs[2].Bytes, _ = hex.DecodeString("2d98021e3cf71580102224b2fcb4c5c60595e8fdf6fd1b97c6ef63e9fb3ed635")
//	e.ExtIDs[3].Bytes, _ = hex.DecodeString("c103756200d0c1223c0ee9911196bf06de6ee570b0f45897e2ef39f9abf39d24")
//	e.ExtIDs[4].Bytes, _ = hex.DecodeString("e56926c84f7edb0774539e03aa4d16e8220c4f3746c872feb4afcc52abf06cceb61cf417c8af417699fc0171bdf9bd543219a23a845776bd6b30c315bd66f100")
//
//	e.Content.Bytes = []byte(`{"proposal":{"title":"Vegetable Award","text":"Please vote for your favorite vegetable"},"vote":{"phasesBlockHeights":{"commitStart":49595,"commitEnd":49595,"revealStart":49596,"revealEnd":49596},"type":1,"config":{"options":["broccoli","spinach","avocado"],"minOptions":1,"maxOptions":2,"acceptanceCriteria":{"minTurnout":{"weighted":0.3,"unweighted":0.5}},"allowAbstention":true,"computeResultsAgainst":"ALL_ELIGIBLE_VOTERS","winnerCriteria":{"minSupport":{"*":{"weighted":0.6,"unweighted":0.4}}}},"eligibleVotersChainId":"84444341e0e60a496f75c98c57357805ec86e9f8e232348f1e60704e83bca2b0"}}`)
//
//	_, err := NewProposalEntry(e, 0)
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestEligibleVoter(t *testing.T) {
	data := `[{"voterId":"2d98021e3cf71580102224b2fcb4c5c60595e8fdf6fd1b97c6ef63e9fb3ed635","weight":2},{"voterId":"44dc565dd5330aaec455583372b233bd1171af531d5083b6d4128b7909218319","weight":6}]`

	var voters []EligibleVoter
	err := json.Unmarshal([]byte(data), &voters)
	if err != nil {
		t.Error(err)
	}
	if len(voters) != 2 {
		t.Errorf("Not all voters marshaled")
	}
}
