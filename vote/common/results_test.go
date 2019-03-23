package common_test

import (
	"testing"

	"fmt"

	"encoding/json"

	. "github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/FactomProject/factomd/common/primitives"
)

type TestVote struct {
	Vote           *Vote
	EligibleVoters []*EligibleVoter
	Reveals        []*VoteReveal
}

func MakeTestVote(options []string, min, max int) *TestVote {
	tv := new(TestVote)

	vote := NewVote()
	vote.Proposal.Vote.Config.Options = options
	vote.Proposal.ProposalChain = primitives.RandomHash()
	vote.Proposal.Vote.Config.ComputeResultsAgainst = "ALL_ELIGIBLE_VOTERS"
	vote.Proposal.Vote.Config.MinOptions = min
	vote.Proposal.Vote.Config.MaxOptions = max

	tv.Vote = vote
	return tv
}

func (tv *TestVote) SetType(t int) {
	tv.Vote.Proposal.Vote.VoteType = t
}

func (tv *TestVote) SetOptions(vector VoteVector) {
	tv.SetType(vector.VoteType)

	if vector.ExtraConfigs != nil {
		opts := vector.ExtraConfigs.opts
		if opt, ok := opts["min"]; ok {
			tv.Vote.Proposal.Vote.Config.MinOptions = opt.(int)
		}
		if opt, ok := opts["max"]; ok {
			tv.Vote.Proposal.Vote.Config.MaxOptions = opt.(int)
		}
		if opt, ok := opts["cpa"]; ok {
			tv.Vote.Proposal.Vote.Config.ComputeResultsAgainst = opt.(string)
		}
		if opt, ok := opts["abs"]; ok {
			tv.Vote.Proposal.Vote.Config.AllowAbstention = opt.(bool)
		}
		if opt, ok := opts["win"]; ok {
			tv.Vote.Proposal.Vote.Config.WinnerCriteria = opt.(WinnerCriteriaStruct)
		}
		if opt, ok := opts["sup"]; ok {
			tv.Vote.Proposal.Vote.Config.AcceptanceCriteria = opt.(AcceptCriteriaStruct)
		}
	}
}

func (tv *TestVote) AddVote(options []string, weight float64) {
	reveal := NewVoteReveal()
	reveal.Content.VoteOptions = options
	reveal.VoteChain = tv.Vote.Proposal.ProposalChain
	reveal.VoterID = primitives.RandomHash()
	reveal.EntryHash = primitives.RandomHash()

	voter := new(EligibleVoter)
	voter.VoteWeight = weight
	voter.VoterID = *(reveal.VoterID.(*primitives.Hash))

	tv.Reveals = append(tv.Reveals, reveal)
	tv.EligibleVoters = append(tv.EligibleVoters, voter)
}

func ExpectedWinners(stats *VoteStats, wins []string) error {
	winners := make(map[string]bool)
	for _, w := range wins {
		winners[w] = true
	}

	foundWins := []string{}
	for _, w := range stats.WeightedWinners {
		foundWins = append(foundWins, w.Option)
	}

	for _, w := range stats.WeightedWinners {
		if _, ok := winners[w.Option]; !ok {
			return fmt.Errorf("Winners incorrect. Exp: %v, found %v", wins, foundWins)
		}
		delete(winners, w.Option)
	}

	if len(winners) > 0 {
		return fmt.Errorf("Winners incorrect. Missing %d. Exp: %v, found %v", len(winners), wins, foundWins)
	}
	return nil
}

func ExpectedChecks(stats *VoteStats, vector VoteVector) error {
	checks := vector.ExtraChecks
	if checks == nil {
		return nil
	}
	// Regular Check
	if checks.OptionStats != nil {
		for k, v := range checks.OptionStats {
			if v2, ok := stats.OptionStats[k]; ok {
				if !v.OptionStats.IsSameAs(v2.OptionStats) {
					return fmt.Errorf("Option %s is not as expected.\n%v\n%v", k, v, v2)
				}
			} else {
				if k == "" {
					v2 := stats.AbstainedStats
					if !v.OptionStats.IsSameAs(v2) {
						return fmt.Errorf("Option 'Abstain' is not as expected.\n%v\n%v", v, v2)
					}
				} else {
					return fmt.Errorf("Key %s is missing in vote", k)
				}
			}
		}
	}

	ww := make(map[string]VoteOptionStats, 0)
	for _, ws := range stats.WeightedWinners {
		ww[ws.Option] = ws
	}
	// Winner Check
	if checks.WinnerStats != nil {
		for k, v := range checks.WinnerStats {
			if v2, ok := ww[k]; ok {
				if !v.IsSameAs(v2) {
					jv, _ := json.Marshal(v)
					jv2, _ := json.Marshal(v2)
					return fmt.Errorf("VoteOption %s is not as expected.\nExp: %s\nFnd: %s", k, string(jv2), string(jv))
				}
			} else {
				return fmt.Errorf("Key %s is missing in vote", k)
			}
		}
	}

	if checks.Additional != nil {
		if v, ok := checks.Additional["valid"]; ok {
			if stats.Valid != v.(bool) {
				return fmt.Errorf("Vote valid exp %t, found %t", v.(bool), stats.Valid)
			}
		}
	}
	return nil
}

func (tv *TestVote) Params() (*Vote, []*EligibleVoter, []*VoteReveal) {
	return tv.Vote, tv.EligibleVoters, tv.Reveals
}

// Test all vote Types

func TestVoteVector(t *testing.T) {
	for i, v := range VoteVectors {
		vote := MakeTestVote(v.Options, 1, 10)
		vote.SetOptions(v)
		for _, r := range v.Votes {
			vote.AddVote(r.Options, r.Weight)
		}

		// Results
		stats, err := ComputeResult(vote.Params())
		if err != nil {
			t.Error(err)
		}

		// Check if winners correct
		err = ExpectedWinners(stats, v.Winners)
		if err != nil {
			t.Error(fmt.Errorf("Vect %d (%s) -> %s", i, v.Title, err.Error()))
		}

		err = ExpectedChecks(stats, v)
		if err != nil {
			t.Error(fmt.Errorf("Vect %d (%s) -> %s", i, v.Title, err.Error()))
		}
	}

	//data, _ := json.Marshal(VoteVectors)
	//fmt.Println(string(data))
}

func findVector(title string) (int, *VoteVector) {
	for i, v := range VoteVectors {
		if v.Title == title {
			return i, &v
		}
	}
	return -1, nil
}

func TestSpecificVoteVector(t *testing.T) {
	title := "Abstain Test 1"
	i, vp := findVector(title)
	if vp == nil {
		t.Errorf("Test vector '%s' not found", title)
		t.FailNow()
	}
	v := *vp

	vote := MakeTestVote(v.Options, 1, 10)
	vote.SetOptions(v)
	for _, r := range v.Votes {
		vote.AddVote(r.Options, r.Weight)
	}

	// Results
	stats, err := ComputeResult(vote.Params())
	if err != nil {
		t.Error(err)
	}

	// Check if winners correct
	err = ExpectedWinners(stats, v.Winners)
	if err != nil {
		t.Error(fmt.Errorf("Vect %d (%s) -> %s", i, v.Title, err.Error()))
	}

	err = ExpectedChecks(stats, v)
	if err != nil {
		t.Error(fmt.Errorf("Vect %d (%s) -> %s", i, v.Title, err.Error()))
	}
}

// IRV Voting Tests

func TestIRVVoteSimple(t *testing.T) {
	vote := MakeTestVote([]string{"A", "B", "C"}, 1, 2)
	vote.SetType(VOTE_IRV)
	vote.AddVote([]string{"A"}, 1)

	stats, err := ComputeResult(vote.Params())
	if err != nil {
		t.Error(err)
	}

	err = ExpectedWinners(stats, []string{"A"})
	if err != nil {
		t.Error(err)
	}
}

func TestIRVVoteMultiple(t *testing.T) {
	vote := MakeTestVote([]string{"A", "B", "C"}, 1, 2)
	vote.SetType(VOTE_IRV)
	vote.AddVote([]string{"B", "C"}, 1)
	vote.AddVote([]string{"A", "B"}, 1)
	vote.AddVote([]string{"B"}, 1)
	vote.AddVote([]string{"C"}, 1)
	vote.AddVote([]string{"C"}, 1)

	stats, err := ComputeResult(vote.Params())
	if err != nil {
		t.Error(err)
	}

	err = ExpectedWinners(stats, []string{"B"})
	if err != nil {
		t.Error(err)
	}
}

// Test Vectors

type IndvVote struct {
	Options []string
	Weight  float64
}

type VoteVector struct {
	Title    string     `json:"title,omitempty"`
	VoteType int        `json:"vote_type,omitempty"`
	Options  []string   `json:"options,omitempty"`
	Votes    []IndvVote `json:"votes,omitempty"`
	Weights  []float64  `json:"weights,omitempty"`
	Winners  []string   `json:"winners,omitempty"`

	ExtraConfigs *ExtraConfigs `json:"extra_config,omitempty"`
	ExtraChecks  *ExtraChecks  `json:"extra_checks,omitempty"`
}

type ExtraConfigs struct {
	MinVotes              int                    `json:"min_votes,omitempty"`
	MaxVotes              int                    `json:"max_votes,omitempty"`
	ComputeResultsAgainst string                 `json:"compute_results_against,omitempty"`
	opts                  map[string]interface{} `json:"opts,omitempty"`
}

func NewExtraConfigs(opts map[string]interface{}) *ExtraConfigs {
	c := new(ExtraConfigs)
	c.opts = opts

	return c
}

type ExtraChecks struct {
	OptionStats map[string]VoteOptionStats
	WinnerStats map[string]VoteOptionStats
	Additional  map[string]interface{}
}

/*
type VoteOptionStats struct {
	Option string  `json:"option,omitempty"`
	Count  float64 `json:"count"`
	Weight float64 `json:"weight"`
	OptionStats

	Support         float64 `json:"support"`
	WeightedSupport float64 `json:"weightedSupport"`
}

*/

var VoteVectors = []VoteVector{
	/*******************
	 *	Single Votes   *
	 *******************/

	// Basic
	VoteVector{
		VoteType: VOTE_SINGLE,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 1,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		Winners: []string{"A"},
	},
	// Not enough support
	VoteVector{
		VoteType: VOTE_SINGLE,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 1, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, .5}},
			},
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B", "C"}, 1},
			IndvVote{[]string{"B", "C"}, 1},
		},
		Winners: []string{},
	},

	// Enough support because invalid votes
	VoteVector{
		VoteType: VOTE_SINGLE,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 1, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, .5}},
			},
			"cpa": "PARTICIPANTS_ONLY",
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B", "C"}, 1},
			IndvVote{[]string{"B", "C"}, 1},
		},
		Winners: []string{"A"},
		ExtraChecks: &ExtraChecks{
			WinnerStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 2, Weight: 2}, Support: 2.0 / 3.0, WeightedSupport: 2.0 / 3.0},
			},
		},
	},

	// Abstain wins
	VoteVector{
		VoteType: VOTE_SINGLE,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 1, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, .5}},
			},
			"cpa": "PARTICIPANTS_ONLY",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 2, Weight: 2}},
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 3, Weight: 3}},
			},
		},
		Winners: []string{},
	},
	/*******************
	 * Approval Votes  *
	 *******************/
	VoteVector{
		VoteType: VOTE_SINGLE,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 5, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, .5}},
			},
			"cpa": "PARTICIPANTS_ONLY",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "B", "C"}, 1},
			IndvVote{[]string{"A", "B", "C"}, 1},
			IndvVote{[]string{"C"}, 0.1},
			IndvVote{[]string{"A", "B", "C"}, 1},
			IndvVote{[]string{"A", "B", "C"}, 1},
			IndvVote{[]string{}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 4, Weight: 4}},
				"B": VoteOptionStats{
					OptionStats: OptionStats{Option: "B", Count: 4, Weight: 4}},
				"C": VoteOptionStats{
					OptionStats: OptionStats{Option: "C", Count: 5, Weight: 4.1}},
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 1, Weight: 1}},
			},
			WinnerStats: map[string]VoteOptionStats{
				"C": VoteOptionStats{
					OptionStats: OptionStats{Option: "C", Count: 5, Weight: 4.1}, Support: 5.0 / 6.0, WeightedSupport: 4.1 / 5.1},
			},
		},
		Winners: []string{"C"},
	},

	/*******************
	*	Binary Votes   *
	********************/

	// Abstain wins because of weight, not count
	VoteVector{
		VoteType: VOTE_BINARY,
		Options:  []string{"A", "B"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 5, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, 0.5}},
			},
			"cpa": "PARTICIPANTS_ONLY",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{}, 10},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 4, Weight: 4}},
				"B": VoteOptionStats{
					OptionStats: OptionStats{Option: "B", Count: 1, Weight: 1}},
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 1, Weight: 10}},
			},
		},
		Winners: []string{},
	},

	// Edge out the victory by 0.01
	VoteVector{
		VoteType: VOTE_BINARY,
		Options:  []string{"A", "B"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 2, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.3, 0.5}},
			},
			"cpa": "PARTICIPANTS_ONLY",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1.01},
			IndvVote{[]string{"B"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 1, Weight: 1.01}},
				"B": VoteOptionStats{
					OptionStats: OptionStats{Option: "B", Count: 1, Weight: 1}},
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 0, Weight: 0}},
			},
		},
		Winners: []string{"A"},
	},

	// No winner
	VoteVector{
		VoteType: VOTE_BINARY,
		Options:  []string{"A", "B"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 2, "win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.3, 0.5}},
			},
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1.01},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"C"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 1, Weight: 1.01}},
				"B": VoteOptionStats{
					OptionStats: OptionStats{Option: "B", Count: 1, Weight: 1}},
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 0, Weight: 0}},
			},
		},
		Winners: []string{},
	},

	/****************
	 *	IRV Votes   *
	 ****************/

	//// Simple winner
	VoteVector{
		VoteType: VOTE_IRV,
		Options:  []string{"A", "B", "C"},
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		Winners: []string{"A"},
	},
	// 2nd level winner -- Edge case
	VoteVector{
		VoteType: VOTE_IRV,
		Options:  []string{"A", "B", "C"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C"}, 1},
		},
		Winners: []string{},
	},
	VoteVector{
		VoteType: VOTE_IRV,
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "B", "D"}, 1},
			IndvVote{[]string{"B", "A", "D"}, 1},
			IndvVote{[]string{"D"}, 1},
			IndvVote{[]string{"D"}, 1},
			IndvVote{[]string{"C"}, 1},
		},
		Winners: []string{"D"},
	},
	// JS Vector
	VoteVector{
		VoteType: VOTE_IRV,
		Options:  []string{"Bob", "Sue", "Bill", "Paul"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"Bob", "Bill", "Sue"}, 1},
			IndvVote{[]string{"Sue", "Bill", "Bob"}, 1},
			IndvVote{[]string{"Paul", "Bill", "Sue"}, 1},
			IndvVote{[]string{"Bob", "Bill", "Sue"}, 1},
			IndvVote{[]string{"Sue", "Bob", "Bill"}, 1},
			IndvVote{[]string{"Sue", "Bill", "Bob"}, 1},
		},
		Winners: []string{"Sue"},
	},
	// Egs
	VoteVector{
		VoteType: VOTE_IRV,
		Options:  []string{"No", "Maybe"},
		Votes: []IndvVote{
			IndvVote{[]string{"No", "Maybe"}, 1},
		},
		Winners: []string{"No"},
	},
	VoteVector{VoteType: VOTE_IRV,
		Options: []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
		},
		Winners: []string{},
	},

	//* Aproval Votes *

	VoteVector{
		VoteType: VOTE_SINGLE,
		Title:    "Ben Set Approval 0",
		Options:  []string{"A", "B", "C", "D", "E"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "B", "C", "E"}, 3},
			IndvVote{[]string{"A", "B", "C"}, 3},
			IndvVote{[]string{"A", "B", "C"}, 3},
			IndvVote{[]string{"E", "D"}, 7},
			IndvVote{[]string{"D"}, 3.01},
		},

		Winners: []string{"D"},
	},

	VoteVector{
		VoteType: VOTE_SINGLE,
		Title:    "Ben Set Approval 1",
		Options:  []string{"A", "B", "C", "D", "E"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
			"win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, 0.5}},
			},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "B", "C", "E"}, 3},
			IndvVote{[]string{"A", "B", "C"}, 3},
			IndvVote{[]string{"A", "B", "C"}, 3},
			IndvVote{[]string{"E", "D"}, 7},
			IndvVote{[]string{"D"}, 3.01},
			IndvVote{[]string{}, 4},
		},

		Winners: []string{},
	},

	//* IRV Votes *

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 1",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "B", "C", "D"}, 1},
			IndvVote{[]string{"B", "C", "D", "A"}, 1},
			IndvVote{[]string{"C", "D", "A", "B"}, 1},
			IndvVote{[]string{"D", "A", "B", "C"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 2",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 3",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
			IndvVote{[]string{"A", "B"}, 1},
			IndvVote{[]string{"D", "B"}, 1},
		},
		Winners: []string{"B"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 4",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
			IndvVote{[]string{"A", "B"}, 1},
			IndvVote{[]string{"D", "C"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 5",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
			IndvVote{[]string{"A", "D"}, 1},
			IndvVote{[]string{"D", "C"}, 1},
		},
		Winners: []string{"C"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 6",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "C", "D", "E"}, 1},
			IndvVote{[]string{"A", "C", "D", "E"}, 1},
			IndvVote{[]string{"B", "C", "D", "E"}, 1},
			IndvVote{[]string{"B", "C", "D", "E"}, 1},
			IndvVote{[]string{"C", "E", "F", "A"}, 1},
			IndvVote{[]string{"D", "E", "F", "A"}, 1},
		},
		Winners: []string{"A"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 7",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"B", "D", "E", "F", "A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"D", "A"}, 1},
		},
		Winners: []string{"A"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 8",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C", "D", "E", "F", "A"}, 1},
			IndvVote{[]string{"D"}, 1},
		},
		Winners: []string{"A"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 10",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "E", "F"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C", "E", "F"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"D"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 11",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"B", "D"}, 1},
			IndvVote{[]string{"B", "A"}, 1},
			IndvVote{[]string{"C", "B"}, 1},
			IndvVote{[]string{"C", "D"}, 1},
			IndvVote{[]string{}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 12",
		Options:  []string{"A", "B", "C", "D"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		Winners: []string{"A"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 13",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"B", "D", "E", "F", "A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"D", "C"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 14",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"B", "D", "E", "F", "A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"C"}, 1},
			IndvVote{[]string{"D", "F"}, 1},
		},
		Winners: []string{},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 15",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "E", "D"}, 1},
			IndvVote{[]string{"A", "E", "C"}, 1},
			IndvVote{[]string{"A", "E", "B"}, 1},
			IndvVote{[]string{"A", "E", "B"}, 1},
			IndvVote{[]string{"A", "E", "C"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"C", "B", "F"}, 1},
			IndvVote{[]string{"C", "B", "F"}, 1},
			IndvVote{[]string{"C", "D", "B"}, 1},
			IndvVote{[]string{"E", "F", "D"}, 1},
			IndvVote{[]string{"F", "B", "C"}, 1},
			IndvVote{[]string{"F", "C", "B"}, 1},
		},
		Winners: []string{"A"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 16",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "E", "C"}, 1},
			IndvVote{[]string{"A", "E", "B"}, 1},
			IndvVote{[]string{"A", "E", "B"}, 1},
			IndvVote{[]string{"A", "E", "C"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"C", "B", "F"}, 1},
			IndvVote{[]string{"C", "B", "D"}, 1},
			IndvVote{[]string{"C", "D", "B"}, 1},
			IndvVote{[]string{"E", "C", "F"}, 1},
			IndvVote{[]string{"F", "B", "C"}, 1},
			IndvVote{[]string{"F", "B", "C"}, 1},
		},
		Winners: []string{"C"},
	},

	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Ben Set 17",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"C", "D", "E"}, 1},
			IndvVote{[]string{"C", "D", "E"}, 1},
			IndvVote{[]string{"C", "D", "E"}, 1},
			IndvVote{[]string{"E", "D", "F"}, 1},
			IndvVote{[]string{"F", "D", "E"}, 1},
			IndvVote{[]string{"F", "D", "E"}, 1},
		},
		Winners: []string{"A"},
	},

	// Abstain
	// Check abstain count, no winner as winner support is needed at 50%
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Abstain Test 0",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
			"sup": AcceptCriteriaStruct{MinTurnout: CriteriaWeights{Unweighted: 0.5}},
			"win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, 0}},
			},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 3, Weight: 3}},
			},
			Additional: map[string]interface{}{
				"valid": true,
			},
		},
		Winners: []string{},
	},

	// Testing the counts. A wins as winner criteria set to 0%
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Abstain Test 0",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 3, "max": 3,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
			"sup": AcceptCriteriaStruct{MinTurnout: CriteriaWeights{Unweighted: 0.5}},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{}, 1},
			IndvVote{[]string{"A", "B", "C"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 3, Weight: 3}},
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 1, Weight: 1}, Support: 1.0 / 4.0, WeightedSupport: 1.0 / 4.0},
			},
			Additional: map[string]interface{}{
				"valid": true,
			},
		},
		Winners: []string{"A"},
	},

	// Check the 1 abstain vote is counted
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Abstain Test 2",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{}, 1},
			IndvVote{[]string{"A"}, 1},
			IndvVote{[]string{"B"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 1, Weight: 1}},
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 2, Weight: 2}, Support: 2.0 / 4.0, WeightedSupport: 2.0 / 4.0},
			},
			Additional: map[string]interface{}{
				"valid": true,
			},
		},
		Winners: []string{"A"},
	},

	// Winner Criteria is 0%, so the highest win, but invalid because not 50% turnout
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Invalid Test 0",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
			"sup": AcceptCriteriaStruct{MinTurnout: CriteriaWeights{Unweighted: 0.5}},
			"win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{0, 0}},
			},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 0, Weight: 0}},
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 1, Weight: 1}, Support: 1.0 / 4.0, WeightedSupport: 1.0 / 4.0},
			},
			Additional: map[string]interface{}{
				"valid": false,
			},
		},
		Winners: []string{"A"},
	},

	// PARTICIPANTS_ONLY so A wins
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Invalid Test 1",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "PARTICIPANTS_ONLY",
			"abs": true,
			"sup": AcceptCriteriaStruct{MinTurnout: CriteriaWeights{Unweighted: 0.0}},
			"win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, 0}},
			},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 0, Weight: 0}},
				"A": VoteOptionStats{
					OptionStats: OptionStats{Option: "A", Count: 1, Weight: 1}, Support: 1.0 / 1.0, WeightedSupport: 1.0 / 1.0},
			},
			Additional: map[string]interface{}{
				"valid": true,
			},
		},
		Winners: []string{"A"},
	},

	// Acceptance Criteria set to 50%
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Invalid Test 2",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
			"sup": AcceptCriteriaStruct{MinTurnout: CriteriaWeights{Unweighted: 0.5}},
			"win": WinnerCriteriaStruct{
				MinSupport: map[string]CriteriaWeights{"*": CriteriaWeights{.5, 0}},
			},
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"Invalid"}, 1},
			IndvVote{[]string{"A"}, 1},
		},
		ExtraChecks: &ExtraChecks{
			OptionStats: map[string]VoteOptionStats{
				"": VoteOptionStats{
					OptionStats: OptionStats{Option: "", Count: 0, Weight: 0}},
			},
			Additional: map[string]interface{}{
				"valid": false,
			},
		},
		Winners: []string{},
	},

	// Non-Specific
	VoteVector{
		VoteType: VOTE_IRV,
		Title:    "Extra Set(0) 1",
		Options:  []string{"A", "B", "C", "D", "E", "F"},
		ExtraConfigs: NewExtraConfigs(map[string]interface{}{
			"min": 1, "max": 10,
			"cpa": "ALL_ELIGIBLE_VOTERS",
			"abs": true,
		}),
		Votes: []IndvVote{
			IndvVote{[]string{"F", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"A", "D", "E"}, 1},
			IndvVote{[]string{"C", "D", "A"}, 1},
			IndvVote{[]string{"C", "D", "A"}, 1},
			IndvVote{[]string{"E", "D", "A"}, 1},
			IndvVote{[]string{"F", "D", "E"}, 1},
			IndvVote{[]string{"F", "D", "E"}, 1},
		},
		Winners: []string{"A"},
		ExtraChecks: &ExtraChecks{
			Additional: map[string]interface{}{
				"valid": true,
			},
		},
	},
}
