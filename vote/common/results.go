package common

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

var (
	VOTE_BINARY = 0
	VOTE_SINGLE = 1
	VOTE_IRV    = 2
)

var plog = log.WithField("file", "results")

// This code is structured similar to the Javascript implementation:
// https://github.com/PaulBernier/factom-vote/blob/master/src/read-vote/compute-vote-result.js

func ComputeResult(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) (*VoteStats, error) {
	reveals = FilterInvalidVotes(vote, eligibleVoters, reveals)
	switch vote.Proposal.Vote.VoteType {
	case VOTE_BINARY: // Binary :: 2 options, no abstain
		return ComputeBinaryVote(vote, eligibleVoters, reveals)
	case VOTE_SINGLE: // Single Option
		return ComputeSingleVote(vote, eligibleVoters, reveals)
	case VOTE_IRV: // Instant Run-Off Voting
		return ComputeIRVVote(vote, eligibleVoters, reveals)
	}
	return nil, fmt.Errorf("unsupported vote type: %d", vote.Proposal.Vote.VoteType)
}

func FilterInvalidVotes(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) []*VoteReveal {
	if vote.Proposal.ProposalChain.String() == "677a0f58308d5b9e7c31b05843e976399c233e0625a11a940605b250945e459c" {
		fmt.Println("FOUND")
	}
	minOptions := vote.Proposal.Vote.Config.MinOptions
	maxOptions := vote.Proposal.Vote.Config.MaxOptions

	// First convert the eligible voters to a map. We will remove from the map as we count the votes.
	voterMap := make(map[[32]byte]*EligibleVoter)
	for _, v := range eligibleVoters {
		voterMap[v.VoterID.Fixed()] = v
	}

	validOptions := make(map[string]bool)
	for _, s := range vote.Proposal.Vote.Config.Options {
		validOptions[s] = true
	}

	var validVotes []*VoteReveal
	for _, r := range reveals {
		// Voter exists
		if _, ok := voterMap[r.VoterID.Fixed()]; ok {
			// Option length is ok
			if len(r.Content.VoteOptions) >= minOptions && len(r.Content.VoteOptions) <= maxOptions {
				for _, v := range r.Content.VoteOptions {
					// All vote options exist
					if _, ok := validOptions[v]; !ok && v != "" {
						continue
					}
				}
				// Valid
				validVotes = append(validVotes, r)
				delete(voterMap, r.VoterID.Fixed())
			} else if len(r.Content.VoteOptions) == 0 && vote.Proposal.Vote.Config.AllowAbstention {
				// Valid
				validVotes = append(validVotes, r)
				delete(voterMap, r.VoterID.Fixed())
			}
		}
	}
	if vote.Proposal.ProposalChain.String() == "677a0f58308d5b9e7c31b05843e976399c233e0625a11a940605b250945e459c" {
		fmt.Println("FOUND")
	}
	return validVotes
}

type IRVRoundResult struct {
	Option string
	Count  float64
	Weight float64
}

func ComputeIRVVote(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) (*VoteStats, error) {
	flog := plog.WithFields(log.Fields{"vote": vote.Proposal.ProposalChain.String(), "func": "ComputeIRVVote"})

	var availableOptions = make(map[string]bool)
	stats := NewVoteStats()
	stats.VoteChain = vote.Proposal.ProposalChain.String()
	for _, opt := range vote.Proposal.Vote.Config.Options {
		var o VoteOptionStats
		o.Option = opt
		stats.OptionStats[opt] = o
		availableOptions[opt] = true
	}

	// First convert the eligible voters to a map for lookup
	voterMap := make(map[[32]byte]*EligibleVoter)
	for _, v := range eligibleVoters {
		voterMap[v.VoterID.Fixed()] = v
		stats.CompleteStats.Count += 1
		stats.CompleteStats.Weight += float64(v.VoteWeight)
	}

	var revealCopies []*VoteReveal
	for _, r := range reveals {
		revealCopies = append(revealCopies, r.Copy())
	}

	// Some totals
	for _, r := range revealCopies {
		// Get the totals
		if voter, ok := voterMap[r.VoterID.Fixed()]; ok {
			stats.VotedStats.Count += 1
			stats.VotedStats.Weight += float64(voter.VoteWeight)
		}
	}

	var roundResults []map[string]IRVRoundResult
	var winner *IRVRoundResult = nil
	// IRV will continue to eliminate vote options until 1 remains
	for winner == nil { // Rounds
		round := make(map[string]IRVRoundResult)

		// Init rounds to 0 votes per option
		for opt, _ := range availableOptions {
			round[opt] = IRVRoundResult{Option: opt, Count: 0, Weight: 0}
		}

		// Tally votes
		for _, r := range revealCopies {
			if voter, ok := voterMap[r.VoterID.Fixed()]; ok {
				// Grab first available vote (if there is one)
				for _, voteOpt := range r.Content.VoteOptions {
					if _, ok := availableOptions[voteOpt]; ok {
						addRoundVote(round, voteOpt, voter.VoteWeight)
						break
					}
				}
			}
		}

		roundResults = append(roundResults, round)
		if len(round) == 0 {
			// No votes
			flog.Infof("IRV has no winner")
			break
		}

		winner = majority(round)
		if winner == nil {
			losers := minority(round)
			for _, l := range losers {
				delete(availableOptions, l.Option)
			}
		}
	}

	// Set stats to the last round
	stats.IRVRounds = roundResults
	lastRound := roundResults[len(roundResults)-1]
	for opt, res := range lastRound {
		stat := stats.OptionStats[opt]
		stat.Weight = res.Weight
		stat.Count = res.Count
		stats.OptionStats[opt] = stat
	}

	err := stats.ComputeSupport(vote)
	if err != nil {
		return nil, err
	}

	// The winner stat comes from the IRV round, not the computer winners
	// TODO: Does computer winners need to be called for min support?
	if winner != nil {
		stat, ok := stats.OptionStats[winner.Option]
		if ok {
			stats.WeightedWinners = []VoteOptionStats{stat}
		}
	}

	return stats, nil
}

func minority(roundResult map[string]IRVRoundResult) []IRVRoundResult {
	lowest := -1.0
	var lowestR []IRVRoundResult
	for _, r := range roundResult {
		if lowest == -1 {
			lowest = r.Count
			lowestR = append(lowestR, r)
			continue
		}
		if r.Weight < lowest {
			lowest = r.Count
			lowestR = []IRVRoundResult{r}
			continue
		}
		if r.Weight == lowest {
			lowestR = append(lowestR, r)
		}
	}
	return lowestR
}

func majority(roundResult map[string]IRVRoundResult) *IRVRoundResult {
	//minSupport := winnerCrit.MinSupport
	//winners := make([]IRVRoundResult, 0)
	//// Uses minsupport
	//for _, r := range roundResult {
	//	if r.Weight > minSupport[r.Option].Weighted {
	//
	//	}
	//}

	// Uses 50% weight threshold
	total := float64(0)
	for _, r := range roundResult {
		total += r.Count
	}
	threshold := math.Ceil((total / 2) + 0.5)
	for _, r := range roundResult {
		if r.Count >= threshold {
			return &r
		}
	}
	return nil
}

func addRoundVote(round map[string]IRVRoundResult, opt string, weight float64) {
	if res, ok := round[opt]; ok {
		res.Weight += weight
		res.Count += 1
		round[opt] = res
		return
	}

	round[opt] = IRVRoundResult{Option: opt, Count: 1, Weight: weight}
}

func ComputeBinaryVote(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) (*VoteStats, error) {
	// Gather vote stats
	stats, err := ComputeVoteStatistics(vote, eligibleVoters, reveals)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func ComputeSingleVote(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) (*VoteStats, error) {
	// Gather vote stats
	stats, err := ComputeVoteStatistics(vote, eligibleVoters, reveals)
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

func (a OptionStats) IsSameAs(b OptionStats) bool {
	if a.Option != b.Option {
		return false
	}
	if a.Count != b.Count {
		return false
	}
	if a.Weight != b.Weight {
		return false
	}
	return true
}

type VoteOptionStats struct {
	OptionStats
	Support         float64 `json:"support"`
	WeightedSupport float64 `json:"weightedSupport"`
}

func (a VoteOptionStats) IsSameAs(b VoteOptionStats) bool {
	if !a.OptionStats.IsSameAs(b.OptionStats) {
		return false
	}
	if a.Support != b.Support {
		return false
	}
	if a.WeightedSupport != b.WeightedSupport {
		return false
	}
	return true
}

type VoteStats struct {
	VoteChain      string                     `json:"chainId"`
	Valid          bool                       `json:"valid"` // If the vote has hit the acceptance criteria
	InvalidReason  string                     `json:"invalidReason,omitempty"`
	CompleteStats  OptionStats                `json:"total"`   // Total count and weight of all voters
	VotedStats     OptionStats                `json:"voted"`   // Total count and weight of all who voted (not abstained)
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

	IRVRounds []map[string]IRVRoundResult `json:"irvRounds, omitempty"`

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
			if opt.Support >= minSupport.Unweighted && opt.WeightedSupport >= minSupport.Weighted {
				s.WeightedWinners = append(s.WeightedWinners, opt)
			}
		} else if minSupport, ok := criteria.MinSupport["*"]; ok { // All options use this by default if not explicit
			if opt.Support >= minSupport.Unweighted && opt.WeightedSupport >= minSupport.Weighted {
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
func ComputeVoteStatistics(vote *Vote, eligibleVoters []*EligibleVoter, reveals []*VoteReveal) (*VoteStats, error) {
	flog := plog.WithFields(log.Fields{"vote": vote.Proposal.ProposalChain.String()})

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

	// Run through the reveals to tally up the vote options
	for _, r := range reveals {
		if voter, ok := voterMap[r.VoterID.Fixed()]; ok {
			if vote.Proposal.Vote.Config.AllowAbstention && len(r.Content.VoteOptions) == 0 {
				stats.AbstainedStats.Count += 1
				stats.AbstainedStats.Weight += float64(voter.VoteWeight)
			} else if len(r.Content.VoteOptions) > maxOptions || len(r.Content.VoteOptions) < minOptions {
				flog.WithFields(log.Fields{"eHash": r.EntryHash.String(), "reason": "optioncount"}).Errorf("Toss")
				continue // Ignore, as it does not have the correct amount of votes
			}

			for _, v := range r.Content.VoteOptions {
				stat, ok := stats.OptionStats[v]
				if !ok {
					continue
				}
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
