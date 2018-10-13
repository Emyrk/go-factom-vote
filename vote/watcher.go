package vote

import (
	"fmt"

	"sync"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// VoteWatcher watches the blockchain for proposals and votes
type VoteWatcher struct {
	// The map key is the vote-chain
	VoteProposals map[[32]byte]*Vote
	OldEntries    []*OldEntry

	// Eligible Voter Lists
	EligibleLists map[[32]byte]*EligibleList

	sync.RWMutex
}

func NewVoteWatcher() *VoteWatcher {
	vw := new(VoteWatcher)
	vw.VoteProposals = make(map[[32]byte]*Vote)
	vw.EligibleLists = make(map[[32]byte]*EligibleList)

	return vw
}

func (vw *VoteWatcher) AddNewVoteProposal(p *Vote) {
	vw.Lock()
	vw.VoteProposals[p.Proposal.ProposalChain.Fixed()] = p
	vw.Unlock()
}

// ProcessEntry will take an entry and apply it to a vote if one exists for the entry
//
//
//	Returns:
//		bool 	True if a vote was updated or changed
//		error
//
func (vw *VoteWatcher) ProcessEntry(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, error) {
	if entry == nil {
		return false, fmt.Errorf("Entry is nil")
	}

	// Not an entry we care about
	if len(entry.ExternalIDs()) < 1 {
		return false, nil
	}

	change, tryagain, err := false, false, error(nil)

	switch string(entry.ExternalIDs()[0]) {
	// First entry to start a vote
	case EXT0_VOTE_CHAIN:
		change, tryagain, err = vw.ProcessVoteChain(entry, dBlockHeight, dBlockTimestamp, newEntry)
	case EXT0_VOTE_COMMIT:
		change, tryagain, err = vw.ProcessVoteCommit(entry, dBlockHeight, dBlockTimestamp, newEntry)
	case EXT0_VOTE_REVEAL:
		change, tryagain, err = vw.ProcessVoteReveal(entry, dBlockHeight, dBlockTimestamp, newEntry)
	case EXT0_VOTE_REGISTRATION_CHAIN:
		// This doesn't need to do anything
	case EXT0_REGISTER_VOTE:
		change, tryagain, err = vw.ProcessVoteRegister(entry, dBlockHeight, dBlockTimestamp, newEntry)
	case EXT0_ELIGIBLE_VOTER_CHAIN:
		change, tryagain, err = vw.ProcessNewEligibleList(entry, dBlockHeight, dBlockTimestamp, newEntry)
		//case
	}

	if err != nil {
		return change, err
	}

	if tryagain && newEntry {
		vw.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
	}

	return change, nil
}

type OldEntry struct {
	Entry           interfaces.IEBEntry
	DBlockHeight    uint32
	DBlockTimestamp uint64
}

func (vw *VoteWatcher) PushEntryForLater(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp) {
	oe := new(OldEntry)
	oe.Entry = entry
	oe.DBlockHeight = dBlockHeight
	oe.DBlockTimestamp = dBlockTimestamp.GetTimeMilliUInt64()

	vw.OldEntries = append(vw.OldEntries, oe)
}

func (vw *VoteWatcher) ProcessOldEntries() (bool, error) {
	var change bool
	for i := 0; i < len(vw.OldEntries); i++ {
		oe := vw.OldEntries[i]
		t := primitives.NewTimestampFromMilliseconds(oe.DBlockTimestamp)
		// Process and Remove
		localchange, _ := vw.ProcessEntry(oe.Entry, oe.DBlockHeight, t, false)
		vw.OldEntries = append(vw.OldEntries[:i], vw.OldEntries[i+1:]...)
		// Set change
		change = change || localchange
	}
	return change, nil
}

/*
 * Different Entry Types
 */

// ProcessVoteChain
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessVoteChain(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	if _, ok := vw.VoteProposals[entry.GetChainID().Fixed()]; ok {
		// Already exists
		return false, true, fmt.Errorf("vote chain already exists")
	}

	v := NewVote()

	// Make the proposal entry
	proposal, err := NewProposalEntry(entry)
	if err != nil {
		return false, false, err
	}

	if valid, err := proposal.IsDataValid(); !valid {
		return false, false, fmt.Errorf("invalid proposal: %s", err.Error())
	}
	v.Proposal = proposal

	list, ok := vw.EligibleLists[v.Proposal.Vote.EligibleVotersChainID.Fixed()]
	if !ok {
		return false, true, fmt.Errorf("no eligible voter list with chain: %s", v.Proposal.Vote.EligibleVotersChainID.String())
	}
	v.EligibleList = list

	vw.AddNewVoteProposal(v)

	return true, false, nil
}

// ProcessVoteCommit
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessVoteCommit(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	//	This entry is in the vote-chain
	v, ok := vw.VoteProposals[entry.GetChainID().Fixed()]
	if !ok {
		return false, true, fmt.Errorf("vote chain does not exist for commit")
	}

	c, err := NewVoteCommit(entry)
	if err != nil {
		return false, false, err
	}

	// We deference, as this structure is now immutable
	err = v.AddCommit(*c, dBlockHeight)
	if err != nil {
		return false, false, err
	}

	return true, false, nil
}

// ProcessVoteReveal
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessVoteReveal(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	//	This entry is in the vote-chain
	v, ok := vw.VoteProposals[entry.GetChainID().Fixed()]
	if !ok {
		return false, true, fmt.Errorf("vote chain does not exist for reveal")
	}

	r, err := NewVoteReveal(entry)
	if err != nil {
		return false, false, err
	}

	// We deference, as this structure is now immutable
	err = v.AddReveal(*r, dBlockHeight)
	if err != nil {
		return false, false, err
	}

	return true, false, nil
}

// ProcessVoteReveal
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessVoteRegister(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	//	This entry is in the vote-chain
	v, ok := vw.VoteProposals[entry.GetChainID().Fixed()]
	if !ok {
		// TODO: Should we allow a register before the vote exists?
		return false, true, fmt.Errorf("vote chain does not exist to be registered")
	}

	v.Registered = true
	return true, false, nil
}

// ProcessNewEligibleList
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessNewEligibleList(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp interfaces.Timestamp,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	//	This entry is in the vote-chain
	_, ok := vw.EligibleLists[entry.GetChainID().Fixed()]
	if ok {
		// Already here
		return false, false, fmt.Errorf("eligibility list already exists")
	}

	list := NewEligibleList()
	head, err := NewEligibleVoterHeader(entry)
	if err != nil {
		return false, false, err
	}
	list.EligibilityHeader = *head
	list.ChainID = entry.GetChainID()
	vw.EligibleLists[entry.GetChainID().Fixed()] = list

	return true, false, nil
}
