package vote

import (
	"fmt"

	"encoding/hex"

	"encoding/json"

	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Vote is the master structure to keep track of an ongoing vote
//
type Vote struct {
	// All votes have a proposal
	Proposal *ProposalEntry

	// Eligible Voters
	Eligibility    EligibleVoterHeader
	EligibleVoters map[[32]byte]EligibleVoter

	// CommitmentPhase
	Commits map[[32]byte]VoteCommit

	// RevealPhas
	Reveals map[[32]byte]VoteReveal

	// Keeps track of the current eblock synced up too
	VoteChainSync *identity.EntryBlockSync
}

func NewVote() *Vote {
	v := new(Vote)
	v.EligibleVoters = make(map[[32]byte]EligibleVoter)
	v.Commits = make(map[[32]byte]VoteCommit)
	v.Reveals = make(map[[32]byte]VoteReveal)
	v.VoteChainSync = identity.NewEntryBlockSync()
	return v
}

// AddVoter will check signatures, and add/remove voters given
//			If the signature is invalid, it will return an error
//	params:
//		e EligibleVoterEntry
//	returns:
//		number of voters applied (added/removed)
//		error if signature is invalid
func (v *Vote) AddVoter(e EligibleVoterEntry, height int) (int, error) {
	if height >= v.Proposal.Vote.PhasesBlockHeights.CommitStart {
		return 0, fmt.Errorf("Vote has already started.")
	}

	// TODO: Check signature
	for _, eg := range e.Content {
		if _, ok := v.EligibleVoters[eg.VoterID.Fixed()]; eg.VoteWeight == 0 && ok {
			delete(v.EligibleVoters, eg.VoterID.Fixed())
		} else {
			v.EligibleVoters[eg.VoterID.Fixed()] = eg
		}
	}

	return 0, nil
}

//
func (v *Vote) AddCommit(c VoteCommit, height uint32) error {
	if int(height) < v.Proposal.Vote.PhasesBlockHeights.CommitStart {
		return fmt.Errorf("commit phase has not started")
	}

	if int(height) > v.Proposal.Vote.PhasesBlockHeights.CommitEnd {
		return fmt.Errorf("commit phase has ended")
	}

	_, ok := v.EligibleVoters[c.VoterID.Fixed()]
	if !ok {
		return fmt.Errorf("not an eligible voter")
	}

	// TODO: Check signature

	// Overwrite vote if already exists
	v.Commits[c.VoterID.Fixed()] = c
	return nil
}

func (v *Vote) AddReveal(r VoteReveal, height int) error {
	if height < v.Proposal.Vote.PhasesBlockHeights.RevealStart {
		return fmt.Errorf("commit phase has not started")
	}

	if height > v.Proposal.Vote.PhasesBlockHeights.RevealEnd {
		return fmt.Errorf("commit phase has ended")
	}

	_, ok := v.EligibleVoters[r.VoterID.Fixed()]
	if !ok {
		return fmt.Errorf("not an eligible voter")
	}

	// If commit does not exist, we discard it
	commit, ok := v.Commits[r.VoterID.Fixed()]
	if !ok {
		return fmt.Errorf("no commit found for this reveal")
	}

	// TODO: Check hmac against commit
	var _ = commit

	// If reveal exists, we discard it
	_, ok = v.Reveals[r.VoterID.Fixed()]
	if !ok {
		return fmt.Errorf("reveal already bound. Only 1 reveal allowed")
	}
	v.Reveals[r.VoterID.Fixed()] = r
	return nil
}

type VoteCommit struct {
	VoterID   interfaces.IHash     `json:"voterID"`
	VoterKey  primitives.PublicKey `json:"voterKey"`
	Signature primitives.Signature `json:"signature"`

	Content struct {
		Commitment string `json:"commitment"`
	}
}

func NewVoteCommit(entry interfaces.IEBEntry) (*VoteCommit, error) {
	if len(entry.ExternalIDs()) != 4 {
		return nil, fmt.Errorf("expected 4 extids, found %d", len(entry.ExternalIDs()))
	}

	c := new(VoteCommit)
	hash, err := primitives.HexToHash(string(entry.ExternalIDs()[1]))
	if err != nil {
		return nil, err
	}
	c.VoterID = hash

	key, err := hex.DecodeString(string(entry.ExternalIDs()[2]))
	if err != nil {
		return nil, err
	}
	c.VoterKey.UnmarshalBinary(key)

	sig, err := hex.DecodeString(string(entry.ExternalIDs()[3]))
	if err != nil {
		return nil, err
	}
	c.Signature.SetSignature(sig)

	err = json.Unmarshal(entry.GetContent(), c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type VoteReveal struct {
	VoterID primitives.Hash `json:"voterID"`

	Content struct {
		VoteOptions []string `json:"vote"`
		// At least 16bytes in hex
		Secret string `json:"secret"`
		// The hash function used to generate the committed HMAC
		// (e.g. md5, sha1, sha256, sha512, etc.)
		HmacAlgo string `json:"hmacAlgo"`
	}
}
