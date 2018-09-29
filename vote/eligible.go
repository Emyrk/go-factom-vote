package vote

import (
	"github.com/FactomProject/factomd/common/primitives"
)

// EligibleVoterHeader is the first entry in the chain
type EligibleVoterHeader struct {
	VoteInitiator primitives.Hash `json:"voteInitiator"`
	// It's not really a hash, but it's 32 bytes
	Nonce              primitives.Hash      `json:"nonce"`
	InitiatorKey       primitives.PublicKey `json:"initiatorKey"`
	InitiatorSignature primitives.Signature `json:"initiatorSignature"`
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

