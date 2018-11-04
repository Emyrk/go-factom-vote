package common_test

import (
	"encoding/json"
	"testing"

	. "github.com/Emyrk/go-factom-vote/vote/common"
)

func TestCommitJson(t *testing.T) {
	j := `
	{
		“commitment”: "random"
	}
	`

	c := new(VoteCommit)
	err := json.Unmarshal([]byte(j), c)
	if err != nil {
		t.Error(err)
	}
}

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

	j = `
{"proposal":{"title":"Is the on-chain voting grant a success?","text":"Please tell if you think this grant was a success","externalRef":{"href":"https://the-voting-grant/vote","hash":{"value":"b950beb02d45cc9c96c198243af669b3dc57d31de994d655f506372fdec3a885","algo":"sha256"}}},"vote":{"commitStartBlockHeight":1000,"commitEndBlockHeight":2000,"revealStartBlockHeight":2001,"revealEndBlockHeight":3000,"participantsChainId":"1891a4fc0feb9a2cce9a384992da69eb032a167ca3fe4545ed22783e49c73321","type":0,"config":{"options":["yes","no","maybe"],"minOptions":1,"maxOptions":1,"acceptanceCriteria":{"minTurnout":0.5,"minSupport":0.6,"weighted":true}},"allowVoteOverwrite":true}}`

	p = new(ProposalEntry)
	err = json.Unmarshal([]byte(j), p)
	if err != nil {
		t.Error(err)
	}

	// TODO: Check unmarshaled values
}
