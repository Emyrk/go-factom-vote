package vote

import (
	"github.com/FactomProject/factomd/common/primitives"
)

// ProposalEntry
type ProposalEntry struct {
	// External IDs
	ProtocolVersion int `json:"version"`
	// Identity Chain
	VoteInitiator      primitives.Hash      `json:"voteInitiator"`
	InitiatorKey       primitives.PublicKey `json:"initiatorKey"`
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`

	// Entry Content
	Proposal ProposalContent `json:"proposal"`
	Vote     VoteContent     `json:"vote"`
}

type ProposalContent struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	ExternalRef struct {
		Href string `json:"href"`
		Hash struct {
			Value primitives.Hash `json:"value"`
			Algo  string          `json:"algo"`
		} `json:"hash"`
	} `json:"externalRef"`
}

type VoteContent struct {
	PhasesBlockHeights struct {
		CommitStart int `json:"commitStart"` // start block for vote commitment phase (inclusive),
		CommitEnd   int `json:"commitEnd"`   // end block for vote commitment phase (inclusive),
		// start block for vote reveal phase; must be after the end block of the commit phase (inclusive),
		RevealStart int `json:"revealStart"`
		RevealEnd   int `json:"revealEnd"` // end block for vote reveal phase (inclusive),
	} `json:"phasesBlockHeights"`

	// ID of the eligible-voters-chain (see Eligible Participants section),
	EligibleVotersChainID primitives.Hash `json:"eligibleVotersChainId"`
	VoteType              string          `json:"type"`
	Config                struct {
		Options []string `json:"options"` // a list of options the voters can choose from,
		// boolean flag for allowing voters to cast an abstained vote
		AllowAbstention bool `json:"allowAbstention"`
		// ComputeResultsAgainst (ALL_ELIGIBLE_VOTERS or PARTICIPANTS_ONLY) specifying the mode of
		// computation of the results
		ComputeResultsAgainst string                   `json:"computeResultsAgainst"`
		MinOptions            int                      `json:"minOptions"`         // min number of options the voter must select,
		MaxOptions            int                      `json:"maxOptions"`         // max number of options the voter can select,
		AcceptanceCriteria    AcceptanceCriteriaStruct `json:"acceptanceCriteria"` // (optional) list of terms for accepting the vot
	} `json:"config"`
}

type AcceptanceCriteriaStruct struct {
	// The strings in the map are the options. "OptionA", etc
	MinSupport map[string]struct {
		Weighted   float64 `json:"weighted"`
		Unweighted float64 `json:"unweighted"`
	} `json:"minSupport"`
	MinTurnout struct {
		Weighted   float64 `json:"weighted"`
		Unweighted float64 `json:"unweighted"`
	} `json:"m inTurnout"`
}

// IsDataValid runs a check on the data to check if it's valid against the rules
func (pe *ProposalEntry) IsDataValid() bool {
	// Cannot have both `text` and `externalRef` field
	if pe.Proposal.Text != "" && pe.Proposal.ExternalRef.Href != "" {
		return false
	}

	return true
}
