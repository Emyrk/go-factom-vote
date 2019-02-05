package common_test

import (
	"testing"

	"fmt"

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

func (tv *TestVote) Params() (*Vote, []*EligibleVoter, []*VoteReveal) {
	return tv.Vote, tv.EligibleVoters, tv.Reveals
}

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

func TestIRVVector(t *testing.T) {
	for _, v := range IRVVectors {
		vote := MakeTestVote(v.Options, 1, 10)
		vote.SetType(VOTE_IRV)
		for _, r := range v.Votes {
			vote.AddVote(r, 1)
		}

		stats, err := ComputeResult(vote.Params())
		if err != nil {
			t.Error(err)
		}

		err = ExpectedWinners(stats, v.Winners)
		if err != nil {
			t.Error(err)
		}
	}
}

// Test Vectors

type IRVVector struct {
	Options []string
	Votes   [][]string
	Winners []string
}

var IRVVectors = []IRVVector{
	//// Simple winner
	IRVVector{
		[]string{"A", "B", "C"},
		[][]string{
			[]string{"A"},
			[]string{"B"},
			[]string{"A"},
		},
		[]string{"A"},
	},
	// 2nd level winner -- Edge case
	IRVVector{
		[]string{"A", "B", "C"},
		[][]string{
			[]string{"A"},
			[]string{"B", "A"},
			[]string{"C"},
		},
		[]string{},
	},
	IRVVector{
		[]string{"A", "B", "C", "D"},
		[][]string{
			[]string{"A", "B", "D"},
			[]string{"B", "A", "D"},
			[]string{"D"},
			[]string{"D"},
			[]string{"C"},
		},
		[]string{"D"},
	},
	// JS Vector
	IRVVector{
		[]string{"Bob", "Sue", "Bill", "Paul"},
		[][]string{
			[]string{"Bob", "Bill", "Sue"},
			[]string{"Sue", "Bill", "Bob"},
			[]string{"Paul", "Bill", "Sue"},
			[]string{"Bob", "Bill", "Sue"},
			[]string{"Sue", "Bob", "Bill"},
			[]string{"Sue", "Bill", "Bob"},
		},
		[]string{"Sue"},
	},
	// Egs
	IRVVector{
		[]string{"No", "Maybe"},
		[][]string{
			[]string{"No", "Maybe"},
		},
		[]string{"No"},
	},
}
