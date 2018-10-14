package vote

import (
	log "github.com/sirupsen/logrus"

	"fmt"

	"github.com/Emyrk/factom-raw"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = log.Error

var IdentityRegisterChain, _ = primitives.HexToHash("888888001750ede0eff4b05f0c3f557890b256450cabbb84cada937f9c258327")

// Controller can search the blockchain for an vote,
// and feed entries into the parses to come up with the state of all votes.
type Controller struct {
	Reader factom_raw.Fetcher
	Parser *VoteWatcher
}

func NewAPIController(apiLocation string) *Controller {
	f := new(Controller)
	f.Reader = factom_raw.NewAPIReader(apiLocation)
	f.Parser = NewVoteWatcher()

	return f
}

// IsWorking will check if the api is working (connection ok)
func (a *Controller) IsWorking() bool {
	_, err := a.Reader.FetchDBlockHead()
	return err == nil
}

//
//func (c *Controller) FindAllIdentities() (map[string]*identity.Identity, error) {
//	// Find all registered identities
//	ids, err := c.parseRegisterChain(true)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, chainID := range ids {
//		c.parseIdentityChain(chainID)
//	}
//
//	humanMap := make(map[string]*identity.Identity)
//	for k, v := range c.Parser.IdentityManager.Identities {
//		if v.IdentityChainID == nil || v.IdentityChainID.IsZero() {
//			continue
//		}
//		humanMap[hex.EncodeToString(k[:])] = v
//	}
//	return humanMap, nil
//}

//func (c *Controller) parseRegisterChain(getHashes bool) ([]interfaces.IHash, error) {
//	regEntries, err := c.FetchChainEntriesInCreateOrder(IdentityRegisterChain)
//	if err != nil {
//		return nil, err
//	}
//
//	var ids []interfaces.IHash
//	if getHashes {
//		for _, e := range regEntries {
//			rfi := new(identityEntries.RegisterFactomIdentityStructure)
//			err := rfi.DecodeFromExtIDs(e.Entry.ExternalIDs())
//			if err != nil {
//				continue
//			}
//
//			ids = append(ids, rfi.IdentityChainID)
//		}
//	}
//
//	err = c.Parser.ParseEntryList(regEntries)
//	if err != nil {
//		return nil, err
//	}
//	return ids, nil
//}

func (c *Controller) FindVote(votechain interfaces.IHash) (*Vote, error) {

	err := c.parseVoteChain(votechain)
	if err != nil {
		return nil, err
	}

	return c.Parser.VoteProposals[votechain.Fixed()], nil
}

// FindIdentity can be given an authority chain ID and build the identity state.
//func (c *Controller) FindIdentity(authorityChain interfaces.IHash) (*identity.Identity, error) {
//	// ** Step 1 **
//	// First we need to determine if the identity is registered. We will have to parse the entire
//	// register chain (TODO: Optimize this)
//	_, err := c.parseRegisterChain(false)
//	if err != nil {
//		return nil, err
//	}
//
//	// ** Step 2 **
//	// Parse the authority chain id,
//	err = c.parseIdentityChain(authorityChain)
//	if err != nil {
//		return nil, err
//	}
//
//	// ** Step 3 **
//	// Return the correct identity
//	return c.Parser.GetIdentity(authorityChain), nil
//}

func (c *Controller) parseVoteChain(votechain interfaces.IHash) error {
	entry, err := c.FetchFirstEntry(votechain)
	if err != nil {
		return fmt.Errorf("fetch first entry: %s", err.Error())
	}

	prop, err := NewProposalEntry(entry.Entry)
	if err != nil {
		return fmt.Errorf("parsing prop: %s", err.Error())
	}

	// Parse the voters
	voterEntries, err := c.FetchChainEntriesInCreateOrder(&prop.Vote.EligibleVotersChainID)
	if err != nil {
		return err
	}

	err = c.Parser.ParseEntryList(voterEntries)
	if err != nil {
		return err
	}

	// Parse the vote
	voteEntries, err := c.FetchChainEntriesInCreateOrder(votechain)
	if err != nil {
		return err
	}

	err = c.Parser.ParseEntryList(voteEntries)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) FetchFirstEntry(chain interfaces.IHash) (ParsingEntry, error) {
	var e ParsingEntry

	head, err := c.Reader.FetchHeadIndexByChainID(chain)
	if err != nil {
		return e, err
	}

	block, err := c.Reader.FetchEBlock(head)
	if err != nil {
		return e, err
	}
	for {
		prev := block.GetHeader().GetPrevKeyMR()
		if prev.IsZero() {
			break
		}
		block, err = c.Reader.FetchEBlock(prev)
		if err != nil {
			return e, err
		}
	}

	entryhash := block.GetEntryHashes()[0]
	entry, err := c.Reader.FetchEntry(entryhash)
	if err != nil {
		return e, err
	}

	dblock, err := c.Reader.FetchDBlockByHeight(block.GetDatabaseHeight())
	if err != nil {
		return e, err
	}
	return ParsingEntry{entry, dblock.GetTimestamp(), block.GetDatabaseHeight()}, nil
}

// FetchChainEntriesInCreateOrder will retrieve all entries in a chain in created order
func (c *Controller) FetchChainEntriesInCreateOrder(chain interfaces.IHash) ([]ParsingEntry, error) {
	head, err := c.Reader.FetchHeadIndexByChainID(chain)
	if err != nil {
		return nil, err
	}

	// Get Eblocks
	var blocks []interfaces.IEntryBlock
	next := head
	for {
		if next.IsZero() {
			break
		}

		// Get the EBlock, and add to list to parse
		block, err := c.Reader.FetchEBlock(next)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)

		next = block.GetHeader().GetPrevKeyMR()
	}

	var entries []ParsingEntry
	// Walk through eblocks in reverse order to get entries
	for i := len(blocks) - 1; i >= 0; i-- {
		eb := blocks[i]

		height := eb.GetDatabaseHeight()
		// Get the timestamp
		dblock, err := c.Reader.FetchDBlockByHeight(height)
		if err != nil {
			return nil, err
		}
		ts := dblock.GetTimestamp()

		ehashes := eb.GetEntryHashes()
		for _, e := range ehashes {
			if e.IsMinuteMarker() {
				continue
			}
			entry, err := c.Reader.FetchEntry(e)
			if err != nil {
				return nil, err
			}

			entries = append(entries, ParsingEntry{entry, ts, height})
		}
	}

	return entries, nil
}
