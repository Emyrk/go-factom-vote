package common

import (
	"fmt"

	"encoding/json"

	"encoding/hex"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// EligibleVoterHeader is the first entry in the chain
type EligibleVoterHeader struct {
	VoteInitiator interfaces.IHash `json:"voteInitiator"`
	// It's not really a hash, but it's 32 bytes
	Nonce        interfaces.IHash     `json:"nonce"`
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

	err = e.VoteInitiator.SetBytes(entry.ExternalIDs()[1])
	if err != nil {
		return nil, err
	}

	e.Nonce = new(primitives.Hash)
	err = e.Nonce.SetBytes(entry.ExternalIDs()[2])
	if err != nil {
		return nil, err
	}

	err = e.InitiatorKey.UnmarshalBinary(entry.ExternalIDs()[3])
	if err != nil {
		return nil, err
	}

	err = e.InitiatorSignature.SetSignature(entry.ExternalIDs()[4])
	if err != nil {
		return nil, err
	}
	e.InitiatorSignature.SetPub(e.InitiatorKey[:])

	signData := computeSha512(append(e.Nonce.Bytes(), entry.GetContent()...))
	if !e.InitiatorSignature.Verify(signData) {
		return nil, fmt.Errorf("signature is not valid")
	}

	return e, nil
}

type EligibleVoterEntry struct {
	Nonce              interfaces.IHash     `json:"nonce"`
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`

	Content []EligibleVoter
}

func NewEligibleVoterEntry(entry interfaces.IEBEntry, blockHeight int, signingkey string) (*EligibleVoterEntry, error) {
	if len(entry.ExternalIDs()) != 3 {
		return nil, fmt.Errorf("expected 3 extids, found %d", len(entry.ExternalIDs()))
	}

	e := new(EligibleVoterEntry)
	e.Nonce = new(primitives.Hash)
	err := e.Nonce.SetBytes(entry.ExternalIDs()[1])
	if err != nil {
		return nil, err
	}

	err = e.InitiatorSignature.SetSignature(entry.ExternalIDs()[2])
	if err != nil {
		return nil, err
	}
	key, _ := hex.DecodeString(signingkey)
	e.InitiatorSignature.SetPub(key)
	signData := computeSha512(append(entry.GetChainID().Bytes(), append(e.Nonce.Bytes(), entry.GetContent()...)...))

	if !e.InitiatorSignature.Verify(signData) {
		return nil, fmt.Errorf("signature is invalid")
	}

	var list []EligibleVoter
	err = json.Unmarshal(entry.GetContent(), &list)
	if err != nil {
		return nil, err
	}
	e.Content = list

	for i := range e.Content {
		e.Content[i].BlockHeight = blockHeight
		e.Content[i].EligibleList.SetBytes(entry.GetChainID().Bytes())
		e.Content[i].EntryHash.SetBytes(entry.GetHash().Bytes())
	}

	return e, nil
}

type EligibleVoter struct {
	// Given by Entry
	VoterID    primitives.Hash `json:"voterId"`
	VoteWeight float64         `json:"weight"`

	// Given by entry context
	BlockHeight  int             `json:"blockHeight"`
	EligibleList primitives.Hash `json:"eligibleList"`
	EntryHash    primitives.Hash `json:"entryHash"`

	// Given by factom-walletd
	SigningKeys []string `json:"keys"`
}

func NewEligibleVoter() *EligibleVoter {
	e := new(EligibleVoter)

	return e
}
