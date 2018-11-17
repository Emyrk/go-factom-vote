package common

import (
	"fmt"
)

// This code is structured similar to the Javascript implementation:
// https://github.com/PaulBernier/factom-vote/blob/master/src/read-vote/compute-vote-result.js

func ComputeResult(vote *Vote, eligibleVoters []*EligibleVoter, commits []*VoteCommit, reveals []*VoteReveal) (*VoteStats, error) {
	switch vote.Proposal.Vote.VoteType {
	case 0: // Binary :: 2 options, no abstain
		return ComputeBinaryVote(vote, eligibleVoters, commits, reveals)
	case 1: // Single Option
		return ComputeSingleVote(vote, eligibleVoters, commits, reveals)
	case 2: // Instant Run-Off Voting
		return nil, fmt.Errorf("voting type 'IRV' not implemented")
	}
	return nil, fmt.Errorf("unsupported vote type: %d", vote.Proposal.Vote.VoteType)
}

func ComputeBinaryVote(vote *Vote, eligibleVoters []*EligibleVoter, commits []*VoteCommit, reveals []*VoteReveal) (*VoteStats, error) {
	// Gather vote stats
	stats, err := ComputeVoteStatistics(vote, eligibleVoters, commits, reveals)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func ComputeSingleVote(vote *Vote, eligibleVoters []*EligibleVoter, commits []*VoteCommit, reveals []*VoteReveal) (*VoteStats, error) {
	// Gather vote stats
	stats, err := ComputeVoteStatistics(vote, eligibleVoters, commits, reveals)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

type OptionStats struct {
	Option string  `json:"option,omitempty"`
	Count  float64 `json:"count"`
	Weight float64 `json:"weight"`
}

type VoteOptionStats struct {
	OptionStats
	Support         float64 `json:"support"`
	WeightedSupport float64 `json:"weightedSupport"`
}

type VoteStats struct {
	VoteChain      string                     `json:"chainId"`
	Valid          bool                       `json:"valid"` // If the vote has hit the acceptance criteria
	InvalidReason  string                     `json:"invalidReason,omitempty"`
	CompleteStats  OptionStats                `json:"total"`   // Total count and weight of all voters
	VotedStats     OptionStats                `json:"voted"` // Total count and weight of all who voted (not abstained)
	AbstainedStats OptionStats                `json:"abstain"` // Total count and weight of all who abstained
	OptionStats    map[string]VoteOptionStats `json:"options"` // Count and weight of each option

	Turnout struct {
		UnweightedTurnout float64 `json:"unweightedTurnout"`
		WeightedTurnout   float64 `json:"weightedTurnout"`
	} `json:"turnout"`

	Support struct {
		CountDenominator  float64 `json:"countDenominator"`
		WeightDenominator float64 `json:"weightDenominator"`
	} `json:"support"`

	WeightedWinners []VoteOptionStats `json:"weightedWinners,omitempty"`
}

func NewVoteStats() *VoteStats {
	vs := new(VoteStats)
	vs.OptionStats = make(map[string]VoteOptionStats)
	return vs
}

// ComputeWinners will compute the highest weighted options
func (s *VoteStats) ComputeWinners(v *Vote) {
	maxWeight := float64(0)
	var winners []VoteOptionStats
	for _, optStats := range s.OptionStats {
		if optStats.Weight > maxWeight {
			winners = []VoteOptionStats{optStats}
			maxWeight = optStats.Weight
		} else if optStats.Weight == maxWeight {
			winners = append(winners, optStats)
		}
	}

	// Check against the minimum support
	criteria := v.Proposal.Vote.Config.WinnerCriteria
	for _, opt := range winners {
		// Check for criteria for this option
		if minSupport, ok := criteria.MinSupport[opt.Option]; ok {
			if opt.Count >= minSupport.Unweighted && opt.Weight >= minSupport.Weighted {
				s.WeightedWinners = append(s.WeightedWinners, opt)
			}
		} else if minSupport, ok := criteria.MinSupport["*"]; ok { // All options use this by default if not explicit
			if opt.Count >= minSupport.Unweighted && opt.Weight >= minSupport.Weighted {
				s.WeightedWinners = append(s.WeightedWinners, opt)
			}
		} else {
			// TODO: Do these count?
			s.WeightedWinners = append(s.WeightedWinners, opt)
		}
	}
}

func (s *VoteStats) ComputeSupport(vote *Vote) error {
	switch vote.Proposal.Vote.Config.ComputeResultsAgainst {
	case "ALL_ELIGIBLE_VOTERS":
		s.Support.CountDenominator = s.CompleteStats.Count
		s.Support.WeightDenominator = s.CompleteStats.Weight
	case "PARTICIPANTS_ONLY":
		if vote.Proposal.Vote.Config.AllowAbstention {
			s.Support.CountDenominator = s.VotedStats.Count
			s.Support.WeightDenominator = s.VotedStats.Weight
		} else {
			s.Support.CountDenominator = s.VotedStats.Count + s.AbstainedStats.Count
			s.Support.WeightDenominator = s.VotedStats.Weight + s.AbstainedStats.Weight
		}
	default:
		return fmt.Errorf("'%s' not a supported 'computeResultsAgainst' value", vote.Proposal.Vote.Config.ComputeResultsAgainst)
	}

	if s.Support.WeightDenominator != 0 && s.Support.CountDenominator != 0 {
		for k, opt := range s.OptionStats {
			opt.Support = opt.Count / s.Support.CountDenominator
			opt.WeightedSupport = opt.Weight / s.Support.WeightDenominator
			s.OptionStats[k] = opt
		}
	}

	// Determine if the vote is valid
	criteria := vote.Proposal.Vote.Config.AcceptanceCriteria
	if s.CompleteStats.Weight == 0 || s.CompleteStats.Count == 0 {
		return nil
	}

	s.Turnout.UnweightedTurnout = s.VotedStats.Count / s.CompleteStats.Count
	s.Turnout.WeightedTurnout = s.VotedStats.Weight / s.CompleteStats.Weight

	if s.Turnout.WeightedTurnout > criteria.MinTurnout.Weighted && s.Turnout.UnweightedTurnout > criteria.MinTurnout.Unweighted {
		s.Valid = true
	}

	return nil
}

// ComputeVoteStatistics
func ComputeVoteStatistics(vote *Vote, eligibleVoters []*EligibleVoter, commits []*VoteCommit, reveals []*VoteReveal) (*VoteStats, error) {
	stats := NewVoteStats()
	stats.VoteChain = vote.Proposal.ProposalChain.String()
	for _, opt := range vote.Proposal.Vote.Config.Options {
		var o VoteOptionStats
		o.Option = opt
		stats.OptionStats[opt] = o
	}

	minOptions := vote.Proposal.Vote.Config.MinOptions
	maxOptions := vote.Proposal.Vote.Config.MaxOptions

	// First convert the eligible voters to a map. We will remove from the map as we count the votes.
	voterMap := make(map[[32]byte]*EligibleVoter)
	for _, v := range eligibleVoters {
		voterMap[v.VoterID.Fixed()] = v
		stats.CompleteStats.Count += 1
		stats.CompleteStats.Weight += float64(v.VoteWeight)
	}

	if vote.Proposal.Vote.Config.AllowAbstention {
		// Run through the commits, and count up the abstains.
		// These commits will not have a reveal, so we need to count them
		// from the commits
		for _, c := range commits {
			if c.Content.Commitment == "" {
				if voter, ok := voterMap[c.VoterID.Fixed()]; ok {
					stats.AbstainedStats.Count += 1
					stats.AbstainedStats.Weight += float64(voter.VoteWeight)
					delete(voterMap, c.VoterID.Fixed())
				}
			}
		}
	}

	// Run through the reveals to tally up the vote options
	for _, r := range reveals {
		if len(r.Content.VoteOptions) > maxOptions || len(r.Content.VoteOptions) < minOptions {
			continue // Ignore, as it does not have the correct amount of votes
		}

		if voter, ok := voterMap[r.VoterID.Fixed()]; ok {
			for _, v := range r.Content.VoteOptions {
				stat := stats.OptionStats[v]
				stat.Weight += float64(voter.VoteWeight)
				stat.Count += 1
				stats.OptionStats[v] = stat
			}
			stats.VotedStats.Count += 1
			stats.VotedStats.Weight += float64(voter.VoteWeight)
			delete(voterMap, r.VoterID.Fixed())
		}
	}

	err := stats.ComputeSupport(vote)
	if err != nil {
		stats.InvalidReason = err.Error()
		return stats, err
	}
	stats.ComputeWinners(vote)

	return stats, nil
}
