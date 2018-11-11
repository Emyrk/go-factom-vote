package common

import (
	"encoding/json"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// ProposalEntry
type ProposalEntry struct {
	// External IDs
	ProtocolVersion int `json:"version"`
	// Identity Chain
	ProposalChain      interfaces.IHash     `json:"vote-chain"`
	VoteInitiator      interfaces.IHash     `json:"voteInitiator"`
	InitiatorKey       primitives.PublicKey `json:"initiatorKey"`
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`

	// Entry Content
	Proposal ProposalContent `json:"proposal"`
	Vote     VoteContent     `json:"vote"`

	// Entry Context
	BlockHeight int             `json:"block_height"`
	EntryHash   primitives.Hash `json:"entry_hash"`
}

func NewProposalEntry(entry interfaces.IEBEntry, dbheight int) (*ProposalEntry, error) {
	if len(entry.ExternalIDs()) != 5 {
		return nil, fmt.Errorf("expected 5 external ids, found %d", len(entry.ExternalIDs()))
	}

	p := new(ProposalEntry)
	p.ProposalChain = entry.GetChainID()
	p.ProtocolVersion = 0 // TODO: Parse protocol version
	p.VoteInitiator = new(primitives.Hash)
	p.VoteInitiator.SetBytes(entry.ExternalIDs()[2]) // = hash
	err := p.InitiatorKey.UnmarshalBinary(entry.ExternalIDs()[3])
	if err != nil {
		return nil, fmt.Errorf("(extid[3]): %s", err.Error())
	}

	err = p.InitiatorSignature.SetSignature(entry.ExternalIDs()[4])
	if err != nil {
		return nil, fmt.Errorf("(extid[4]): %s", err.Error())
	}
	p.InitiatorSignature.SetPub(p.InitiatorKey[:])

	// Validate content
	signData := computeSha512(entry.GetContent())
	if !p.InitiatorSignature.Verify(signData) {
		return nil, fmt.Errorf("Invalid signature on proposal")
	}

	err = json.Unmarshal(entry.GetContent(), p)
	if err != nil {
		return nil, err
	}

	p.EntryHash.SetBytes(entry.GetHash().Bytes())
	p.BlockHeight = dbheight

	return p, nil
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
	VoteType              int             `json:"type"`
	Config                struct {
		Options []string `json:"options"` // a list of options the voters can choose from,
		// boolean flag for allowing voters to cast an abstained vote
		AllowAbstention bool `json:"allowAbstention"`
		// ComputeResultsAgainst (ALL_ELIGIBLE_VOTERS or PARTICIPANTS_ONLY) specifying the mode of
		// computation of the results
		ComputeResultsAgainst string         `json:"computeResultsAgainst"`
		MinOptions            int            `json:"minOptions"`         // min number of options the voter must select,
		MaxOptions            int            `json:"maxOptions"`         // max number of options the voter can select,
		AcceptanceCriteria    CriteriaStruct `json:"acceptanceCriteria"` // (optional) list of terms for accepting the vot
		WinnerCriteria        CriteriaStruct `json:"winnerCriteria"`
	} `json:"config"`
}

type CriteriaStruct struct {
	// The strings in the map are the options. "OptionA", etc
	MinSupport map[string]struct {
		Weighted   float64 `json:"weighted"`
		Unweighted float64 `json:"unweighted"`
	} `json:"minSupport"`
	MinTurnout struct {
		Weighted   float64 `json:"weighted"`
		Unweighted float64 `json:"unweighted"`
	} `json:"minTurnout"`
}

// IsDataValid runs a check on the data to check if it's valid against the rules
func (pe *ProposalEntry) IsDataValid() (bool, error) {
	return true, nil
	// Cannot have both `text` and `externalRef` field
	if pe.Proposal.Text != "" && pe.Proposal.ExternalRef.Href != "" {
		return false, fmt.Errorf("cannot have both 'text' and 'externalRef' fields")
	}

	// TODO: What else makes it invalid?

	return true, nil
}
