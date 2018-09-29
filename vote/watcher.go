package vote

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// VoteWatcher watches the blockchain for proposals and votes
type VoteWatcher struct {
	// The map key is the vote-chain
	VoteProposals map[primitives.Hash]Vote
	OldEntries    []*OldEntry
}

// ProcessEntry will take an entry and apply it to a vote if one exists for the entry
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

	switch string(entry.ExternalIDs()[0]) {
	case "EXT0_VOTE_CHAIN":
	case "EXT0_VOTE_COMMIT":
	case "EXT0_VOTE_REVEAL":
	case "EXT0_VOTE_REGISTRATION_CHAIN":
	case "EXT0_REGISTER_VOTE":
	case "EXT0_ELIGIBLE_VOTER_CHAIN":
	}

	return false, nil
}

type OldEntry struct {
	Entry           interfaces.IEBEntry
	DBlockHeight    uint32
	DBlockTimestamp uint64
}

func (vw *VoteWatcher) PushEntryForLater(entry interfaces.IEBEntry, dBlockHeight uint32, dBlockTimestamp interfaces.Timestamp) error {
	oe := new(OldEntry)
	oe.Entry = entry
	oe.DBlockHeight = dBlockHeight
	oe.DBlockTimestamp = dBlockTimestamp.GetTimeMilliUInt64()

	vw.OldEntries = append(vw.OldEntries, oe)
	return nil
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
