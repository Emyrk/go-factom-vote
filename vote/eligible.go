package vote

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EligibleVoterHeader is the first entry in the chain
type EligibleVoterHeader struct {
	VoteInitiator interfaces.IHash `json:"voteInitiator"`
	// It's not really a hash, but it's 32 bytes
	Nonce        []byte               `json:"nonce"`
	InitiatorKey primitives.PublicKey `json:"initiatorKey"`
	// TODO: Is this an ed25519 sig?
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`
}

func NewEligibleVoterHeader(entry interfaces.IEBEntry) (*EligibleVoterHeader, error) {
	if len(entry.ExternalIDs()) != 5 {
		return nil, fmt.Errorf("expected 5 extids, found %d", len(entry.ExternalIDs()))
	}

	var err error
	e := new(EligibleVoterHeader)
	e.VoteInitiator = new(primitives.Hash)
	//e.VoteInitiator, err = primitives.HexToHash(string(entry.ExternalIDs()[1]))
	//if err != nil {
	//	return nil, err
	//}
	e.VoteInitiator.SetBytes(entry.ExternalIDs()[1])

	e.Nonce = entry.ExternalIDs()[2]
	//e.Nonce, err = primitives.HexToHash(string(entry.ExternalIDs()[2]))
	//if err != nil {
	//	return nil, err
	//}

	//b, err := hex.DecodeString(string(entry.ExternalIDs()[3]))
	//if err != nil {
	//	return nil, err
	//}
	err = e.InitiatorKey.UnmarshalBinary(entry.ExternalIDs()[3])
	if err != nil {
		return nil, err
	}

	//s, err := hex.DecodeString(string(entry.ExternalIDs()[4]))
	//if err != nil {
	//	return nil, err
	//}
	err = e.InitiatorSignature.SetSignature(entry.ExternalIDs()[4])
	if err != nil {
		return nil, err
	}

	return e, nil
}

type EligibleVoterEntry struct {
	Nonce              primitives.Hash      `json:"nonce"`
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`

	Content []EligibleVoter
}

type EligibleVoter struct {
	VoterID    primitives.Hash `json:"voterId"`
	VoteWeight int             `json:"weight"`
}
