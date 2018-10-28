package common

import "github.com/FactomProject/factomd/common/interfaces"

type EligibleList struct {
	ChainID interfaces.IHash `json:"chainid"`

	// Eligible Voters
	EligibilityHeader EligibleVoterHeader        `json:"header"`
	EligibleVoters    map[[32]byte]EligibleVoter `json:"voters"`
	// Able to check if an entry was already processed
	SubmittedEntries map[[32]byte]bool `json:"replay-filter"`
}

func NewEligibleList() *EligibleList {
	e := new(EligibleList)
	e.EligibleVoters = make(map[[32]byte]EligibleVoter)
	e.SubmittedEntries = make(map[[32]byte]bool)
	return e
}

// AddVoter will check signatures, and add/remove voters given
//			If the signature is invalid, it will return an error
//	params:
//		e EligibleVoterEntry
//	returns:
//		number of voters applied (added/removed)
//		error if signature is invalid
func (l *EligibleList) AddVoter(e *EligibleVoterEntry) error {
	// TODO: Check signature
	for _, eg := range e.Content {
		if _, ok := l.EligibleVoters[eg.VoterID.Fixed()]; eg.VoteWeight == 0 && ok {
			delete(l.EligibleVoters, eg.VoterID.Fixed())
		} else {
			l.EligibleVoters[eg.VoterID.Fixed()] = eg
		}
	}

	return nil
}
