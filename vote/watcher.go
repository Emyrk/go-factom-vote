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

	SQLDB           *database.SQLDatabase
	WalletdLocation string
	UseMemory       bool

	sync.RWMutex
}

func NewVoteWatcherWithDB(db *database.SQLDatabase) *VoteWatcher {
	vw := newVoteWatcher()
	vw.SQLDB = db

	return vw
}

func NewVoteWatcher() *VoteWatcher {
	var err error
	vw := newVoteWatcher()

	vw.SQLDB, err = database.InitLocalDB()
	if err != nil {
		panic(err)
	}

	return vw
}

func newVoteWatcher() *VoteWatcher {
	vw := new(VoteWatcher)
	vw.VoteProposals = make(map[[32]byte]*Vote)
	vw.EligibleLists = make(map[[32]byte]*EligibleList)
	vw.WalletdLocation = "localhost:8089"

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
			change, tryagain, err = vw.ProcessNewEligibleVoter(entry, dBlockHeight, dBlockTimestamp, newEntry)
		} else {
			change, tryagain, err = vw.ProcessNewEligibleList(entry, dBlockHeight, dBlockTimestamp, newEntry)
		}
	default:
		return false, nil
	}

	if tryagain && newEntry {
		vw.PushEntryForLater(entry, dBlockHeight, dBlockTimestamp)
	}

	if err != nil {
		return change, err
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
	exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
	if exists {
		return false, false, fmt.Errorf("vote chain already exists: %s", entry.GetChainID().String())
	}

	if err != nil {
		return false, true, err
	}

	v := NewVote()

	// Make the proposal entry
	proposal, err := NewProposalEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, err
	}

	if valid, err := proposal.IsDataValid(); !valid {
		return false, false, fmt.Errorf("invalid proposal: %s", err.Error())
	}
	v.Proposal = proposal

	exists, err = vw.SQLDB.IsEligibleListExist(v.Proposal.Vote.EligibleVotersChainID.String())
	if !exists {
		return false, true, fmt.Errorf("no eligible voter list with chain: %s", v.Proposal.Vote.EligibleVotersChainID.String())
	}

	if err != nil {
		return false, true, err
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

	exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
	if !exists {
		return false, true, fmt.Errorf("vote chain does not exist for commit : %s", entry.GetChainID().String())
	}

	if err != nil {
		return false, true, err
	}

	c, err := NewVoteCommitFromEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, err
	}

	// We deference, as this structure is now immutable
	err = vw.AddCommit(*c, dBlockHeight) // v.AddCommit(*c, dBlockHeight)
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

	exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
	if !exists {
		return false, true, fmt.Errorf("vote chain does not exist for reveal")
	}

	if err != nil {
		return false, true, fmt.Errorf("(reveal:exists) %s", err.Error())
	}

	r, err := NewVoteRevealFromEntry(entry, int(dBlockHeight))
	if err != nil {
		return false, false, fmt.Errorf("(reveal:new) %s", err.Error())
	}

	// We deference, as this structure is now immutable
	// Do signature validation in this function, it will interact with the database
	err = vw.AddReveal(*r, dBlockHeight)
	if err != nil {
		return false, false, fmt.Errorf("(reveal:add) %s", err.Error())
	}

	return true, false, nil
}

// ProcessVoteRegister
//	Returns:
//		bool 	True if a vote was updated or changed
//		bool 	Indicates whether it should be tried again (out of order)
//		error
//
func (vw *VoteWatcher) ProcessVoteRegister(entry interfaces.IEBEntry,
	dBlockHeight uint32,
	dBlockTimestamp time.Time,
	newEntry bool) (bool, bool, error) {

	exists, err := vw.SQLDB.IsVoteExist(entry.GetChainID().String())
	if !exists {
		return false, true, fmt.Errorf("vote chain does not exist to be registered")
	}

	if err != nil {
		return false, true, err
	}

	err = vw.SetRegistered(entry.GetChainID(), true)
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

	exists, err := vw.SQLDB.IsEligibleListExist(entry.GetChainID().String())
	if exists {
		return false, true, fmt.Errorf("eligibility list already exists")
	}

	if err != nil {
		return false, true, err
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
			v.BlockHeight = int(dBlockHeight)
			v.EligibleList.SetBytes(entry.GetChainID().Bytes())
			v.EntryHash.SetBytes(entry.GetHash().Bytes())
			list.EligibleVoters[v.VoterID.Fixed()] = v
		}
	}

	list.EligibilityHeader = *head
	list.ChainID = entry.GetChainID()

	data, err := entry.MarshalBinary()
	if err != nil {
		return false, false, err
	}

	hash := sha256.Sum256(data)

	err = vw.AddNewEligibleList(list, hash)
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

	exists, key, err := vw.SQLDB.IsEligibleListExistWithKey(entry.GetChainID().String())
	if !exists {
		return false, true, fmt.Errorf("eligibility list does not exist")
	}

	if err != nil {
		return false, true, err
	}

	data, err := entry.MarshalBinary()
	if err != nil {
		return false, false, err
	}

	hash := sha256.Sum256(data)
	exists, err = vw.SQLDB.IsRepeatedEntryExists(hex.EncodeToString(hash[:]))
	if exists {
		return false, false, fmt.Errorf("repeated eligible entry tossed")
	}

	if err != nil {
		return false, true, err
	}

	ee, err := NewEligibleVoterEntry(entry, int(dBlockHeight), key)
	if err != nil {
		return false, false, err
	}

	err = vw.AddEligibleVoter(ee, hash)
	if err != nil {
		return false, true, err
	}

	return true, false, nil
}
