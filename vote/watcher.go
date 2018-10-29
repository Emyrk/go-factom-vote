package vote

import (
	"fmt"

	"sync"

	"encoding/json"

	"crypto/sha256"

	"time"

	"encoding/hex"

	. "github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/Emyrk/go-factom-vote/vote/database"
	"github.com/FactomProject/factomd/common/interfaces"
	log "github.com/sirupsen/logrus"
)

var vwLogger = log.WithFields(log.Fields{"struct": "VoteWatcher"})

// VoteWatcher watches the blockchain for proposals and votes
type VoteWatcher struct {
	// The map key is the vote-chain
	VoteProposals map[[32]byte]*Vote
	OldEntries    []*OldEntry

	// Eligible Voter Lists
	EligibleLists map[[32]byte]*EligibleList

	SQLDB     *database.SQLDatabase
	UseMemory bool

	sync.RWMutex
}

func NewVoteWatcher() *VoteWatcher {
	var err error
	vw := new(VoteWatcher)
	vw.VoteProposals = make(map[[32]byte]*Vote)
	vw.EligibleLists = make(map[[32]byte]*EligibleList)
	vw.SQLDB, err = database.InitLocalDB()
	if err != nil {
		panic(err)
	}

	return vw
}

// ParsingEntry is parsable, as it contains all the needed info
type ParsingEntry struct {
	Entry       interfaces.IEBEntry
	Timestamp   time.Time
	BlockHeight uint32
}

func (vw *VoteWatcher) ParseEntryList(list []ParsingEntry) error {
	for _, e := range list {
		_, err := vw.ProcessEntry(e.Entry, e.BlockHeight, e.Timestamp, true)
		if err != nil {
			first := ""
			if len(e.Entry.ExternalIDs()) >= 1 {
				first = string(e.Entry.ExternalIDs()[0])
			}
			vwLogger.WithFields(log.Fields{"func": "ParseEntryList", "chain": e.Entry.GetChainID().String(), "entryhash": e.Entry.GetHash().String(), "[0]": first}).Errorf("Error: %s", err.Error())
		}
	}

	// Parse the remaining
	vw.ProcessOldEntries()
	return nil
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
	dBlockTimestamp time.Time,
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
		if len(entry.ExternalIDs()) == 3 {

		} else {
			change, tryagain, err = vw.ProcessNewEligibleList(entry, dBlockHeight, dBlockTimestamp, newEntry)
		}
	default:
		return false, nil
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
	DBlockTimestamp time.Time
}

func (vw *VoteWatcher) PushEntryForLater(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp time.Time) {
	oe := new(OldEntry)
	oe.Entry = entry
	oe.DBlockHeight = dBlockHeight
	oe.DBlockTimestamp = dBlockTimestamp

	vw.OldEntries = append(vw.OldEntries, oe)
}

func (vw *VoteWatcher) ProcessOldEntries() (bool, error) {
	var change bool
	for i := 0; i < len(vw.OldEntries); i++ {
		oe := vw.OldEntries[i]
		t := oe.DBlockTimestamp
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
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	// Votes are indexed by the chain
	if vw.UseMemory {
		if _, ok := vw.VoteProposals[entry.GetChainID().Fixed()]; ok {
			// Already exists
			return false, true, fmt.Errorf("vote chain already exists")
		}
	} else { // Use Postgres
		exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
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

	if vw.UseMemory {
		list, ok := vw.EligibleLists[v.Proposal.Vote.EligibleVotersChainID.Fixed()]
		if !ok {
			return false, true, fmt.Errorf("no eligible voter list with chain: %s", v.Proposal.Vote.EligibleVotersChainID.String())
		}
		v.EligibleList = list
	} else { // Use Postgres for checks
		exists, err := vw.SQLDB.IsEligibleListExist(v.Proposal.Vote.EligibleVotersChainID.String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

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
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	var v *Vote
	var ok bool

	if vw.UseMemory {
		// Votes are indexed by the chain
		//	This entry is in the vote-chain
		v, ok = vw.VoteProposals[entry.GetChainID().Fixed()]
		if !ok {
			return false, true, fmt.Errorf("vote chain does not exist for commit")
		}
	} else {
		exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

	c, err := NewVoteCommitFromEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, err
	}

	// We deference, as this structure is now immutable
	err = vw.AddCommit(v, *c, dBlockHeight) // v.AddCommit(*c, dBlockHeight)
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
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	var v *Vote
	var ok bool

	if vw.UseMemory {
		// Votes are indexed by the chain
		//	This entry is in the vote-chain
		v, ok = vw.VoteProposals[entry.GetChainID().Fixed()]
		if !ok {
			return false, true, fmt.Errorf("vote chain does not exist for reveal")
		}
	} else {
		exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

	r, err := NewVoteRevealFromEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, err
	}

	// We deference, as this structure is now immutable
	err = vw.AddReveal(v, *r, dBlockHeight)
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
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	var v *Vote
	var ok bool

	if vw.UseMemory {
		// Votes are indexed by the chain
		//	This entry is in the vote-chain
		v, ok = vw.VoteProposals[entry.GetChainID().Fixed()]
		if !ok {
			// TODO: Should we allow a register before the vote exists?
			return false, true, fmt.Errorf("vote chain does not exist to be registered")
		}
	} else {
		exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

	err := vw.SetRegistered(v, true)
	if err != nil {
		return false, false, err
	}
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
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	if vw.UseMemory {
		// Votes are indexed by the chain
		//	This entry is in the vote-chain
		_, ok := vw.EligibleLists[entry.GetChainID().Fixed()]
		if ok {
			// Already here
			return false, false, fmt.Errorf("eligibility list already exists")
		}

	} else {
		exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

	list := NewEligibleList()
	head, err := NewEligibleVoterHeader(entry)
	if err != nil {
		return false, false, err
	}

	// Check if any voters in the content
	var voters []EligibleVoter
	err = json.Unmarshal(entry.GetContent(), &voters)
	if err == nil {
		for _, v := range voters {
			list.EligibleVoters[v.VoterID.Fixed()] = v
		}
	}

	list.EligibilityHeader = *head
	list.ChainID = entry.GetChainID()
	err = vw.AddNewEligibleList(list)
	if err != nil {
		return false, false, err
	}

	return true, false, nil
}

// ProcessNewEligibleVoter
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessNewEligibleVoter(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	var list *EligibleList
	var ok bool
	if vw.UseMemory {
		// Votes are indexed by the chain
		//	This entry is in the vote-chain
		list, ok = vw.EligibleLists[entry.GetChainID().Fixed()]
		if ok {
			// Already here
			return false, false, fmt.Errorf("eligibility list already exists")
		}
	} else {
		exists, err := vw.SQLDB.IsEligibleListExist(entry.GetChainID().String())
		if !exists {
			return false, false, fmt.Errorf("vote chain already exists")
		}

		if err != nil {
			return false, true, err
		}
	}

	data, err := entry.MarshalBinary()
	if err != nil {
		return false, false, err
	}

	hash := sha256.Sum256(data)
	if vw.UseMemory {
		if _, ok := list.SubmittedEntries[hash]; ok {
			return false, false, fmt.Errorf("repeated eligible entry tossed")
		}
	} else {
		exists, err := vw.SQLDB.IsRepeatedEntryExists(hex.EncodeToString(hash[:]))
		if exists {
			return false, false, fmt.Errorf("repeated eligible entry tossed")
		}

		if err != nil {
			return false, true, err
		}
	}

	ee, err := NewEligibleVoterEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, err
	}

	err = vw.AddEligibleVoter(list, ee, hash)
	if err != nil {
		return false, false, err
	}

	return true, false, nil
}
