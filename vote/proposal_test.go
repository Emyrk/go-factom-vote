package vote

import (
	"encoding/json"
	"testing"
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

	// TODO: Check unmarshaled values
}
