package common

import (
	"fmt"

	"encoding/json"

	"bytes"

	"github.com/FactomProject/factomd/common/identity"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Vote is the master structure to keep track of an ongoing vote
//
type Vote struct {
	// Has it been registered?
	Registered bool `json:"registered"`

	// All votes have a proposal
	Proposal *ProposalEntry `json:"proposal"`

	// Eligible Voters
	EligibleList *EligibleList `json:"-"`

	// CommitmentPhase
	Commits map[[32]byte]VoteCommit `json:"-"`

	// RevealPhase
	Reveals map[[32]byte]VoteReveal `json:"-"`

	// Keeps track of the current eblock synced up too
	VoteChainSync *identity.EntryBlockSync `json:"-"`
}

func NewVote() *Vote {
	v := new(Vote)
	v.Commits = make(map[[32]byte]VoteCommit)
	v.Reveals = make(map[[32]byte]VoteReveal)
	v.VoteChainSync = identity.NewEntryBlockSync()
	return v
}

func (v *Vote) String() string {
	var r []VoteReveal
	for _, rm := range v.Reveals {
		r = append(r, rm)
	}

	var c []VoteCommit
	for _, cm := range v.Commits {
		c = append(c, cm)
	}

	data, err := json.Marshal(struct {
		Proposal interface{} `json:"proposal"`
		Commits  interface{} `json:"commits"`
		Reveals  interface{} `json:"reveals"`
	}{
		v,
		c,
		r,
	})
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	json.Indent(&buf, data, "", "\t")

	return string(buf.Bytes())
}

//
func (v *Vote) AddCommit(c VoteCommit, height uint32) error {
	if int(height) < v.Proposal.Vote.PhasesBlockHeights.CommitStart {
		return fmt.Errorf("commit phase has not started")
	}

	if int(height) > v.Proposal.Vote.PhasesBlockHeights.CommitEnd {
		return fmt.Errorf("commit phase has ended")
	}

	_, ok := v.EligibleList.EligibleVoters[c.VoterID.Fixed()]
	if !ok {
		return fmt.Errorf("not an eligible voter")
	}

	// TODO: Check signature

	// Overwrite vote if already exists
	v.Commits[c.VoterID.Fixed()] = c
	return nil
}

func (v *Vote) AddReveal(r VoteReveal, height uint32) error {
	if int(height) < v.Proposal.Vote.PhasesBlockHeights.RevealStart {
		return fmt.Errorf("commit phase has not started")
	}

	if int(height) > v.Proposal.Vote.PhasesBlockHeights.RevealEnd {
		return fmt.Errorf("commit phase has ended")
	}

	_, ok := v.EligibleList.EligibleVoters[r.VoterID.Fixed()]
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
	if ok {
		return fmt.Errorf("reveal already bound. Only 1 reveal allowed")
	}
	v.Reveals[r.VoterID.Fixed()] = r
	return nil
}

type VoteCommit struct {
	VoterID     interfaces.IHash     `json:"voterID"`
	VoterKey    primitives.PublicKey `json:"voterKey"`
	Signature   primitives.Signature `json:"signature"`
	VoteChain   interfaces.IHash     `json:"voteChain"`
	EntryHash   interfaces.IHash     `json:"entryHash"`
	BlockHeight int                  `json:"blockHeight"`

	Content struct {
		Commitment string `json:"commitment"`
	} `json:"content"`
}

func NewVoteCommit() *VoteCommit {
	c := new(VoteCommit)
	c.VoterID = new(primitives.Hash)
	c.VoteChain = new(primitives.Hash)

	return c
}

func NewVoteCommitFromEntry(entry interfaces.IEBEntry, blockHeight int) (*VoteCommit, error) {
	if len(entry.ExternalIDs()) != 4 {
		return nil, fmt.Errorf("expected 4 extids, found %d", len(entry.ExternalIDs()))
	}

	c := new(VoteCommit)
	c.VoterID = new(primitives.Hash)
	c.VoterID.SetBytes(entry.ExternalIDs()[1])
	c.VoterKey.UnmarshalBinary(entry.ExternalIDs()[2])
	c.Signature.SetSignature(entry.ExternalIDs()[3])
	c.VoteChain = entry.GetChainID()
	c.EntryHash = entry.GetHash()
	c.BlockHeight = blockHeight

	err := json.Unmarshal(entry.GetContent(), &c.Content)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type VoteReveal struct {
	VoterID     interfaces.IHash `json:"voterID"`
	VoteChain   interfaces.IHash `json:"voteChain"`
	EntryHash   interfaces.IHash `json:"entryHash"`
	BlockHeight int              `json:"blockHeight"`

	Content struct {
		VoteOptions []string `json:"vote"`
		// At least 16bytes in hex
		Secret string `json:"secret"`
		// The hash function used to generate the committed HMAC
		// (e.g. md5, sha1, sha256, sha512, etc.)
		HmacAlgo string `json:"hmacAlgo"`
	} `json:"content"`
}

func NewVoteReveal() *VoteReveal {
	r := new(VoteReveal)
	r.VoterID = new(primitives.Hash)
	r.VoteChain = new(primitives.Hash)

	return r
}

func NewVoteRevealFromEntry(entry interfaces.IEBEntry, blockHeight int) (*VoteReveal, error) {
	if len(entry.ExternalIDs()) != 2 {
		return nil, fmt.Errorf("expected 2 extids, found %d", len(entry.ExternalIDs()))
	}

	r := new(VoteReveal)
	r.VoterID = new(primitives.Hash)
	r.VoterID.SetBytes(entry.ExternalIDs()[1])

	r.VoteChain = entry.GetChainID()
	r.EntryHash = entry.GetHash()

	err := json.Unmarshal(entry.GetContent(), &r.Content)
	if err != nil {
		return nil, err
	}

	r.BlockHeight = blockHeight

	return r, nil
}
