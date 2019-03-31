package vote

const (
	// External ID 0's
	// The string put in ext[0] to signal the type of entry

	///
	// Entries in vote-chain
	///

	// This is the first entry in the chain to setup a vote
	EXT0_VOTE_CHAIN = "factom-vote"
	// Vote Commit
	EXT0_VOTE_COMMIT = "factom-vote-commit"
	// Vote Reveal
	EXT0_VOTE_REVEAL = "factom-vote-reveal"

	///
	// Entries Registration Chain
	///

	// Vote registaton chain
	EXT0_VOTE_REGISTRATION_CHAIN = "factom-vote-registration"
	// Register a Vote in register chain
	EXT0_REGISTER_VOTE = "Register Factom Vote"

	///
	// Entries in Eligible Voters Chain
	///

	// Chain that lists eligible voters
	EXT0_ELIGIBLE_VOTER_CHAIN = "factom-vote-eligible-voters"
)

const (
	REGISTRATION_CHAIN = "a968e880ee3a7002f25ade15ae36a77c15f4dbc9d8c11fdd5fe86ba6af73a475"
)
