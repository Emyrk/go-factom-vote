package common

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/FactomProject/factomd/common/primitives"
)

func (v *Vote) New() ISQLObject {
	return NewVote()
}

func (v *Vote) Table() string {
	return "proposals"
}

func (v *Vote) InsertFunction() string {
	return "insert_vote"
}

func (v *Vote) ScanRow(row SQLRowWithScan) (*Vote, error) {
	var vi, sigKey, sig, egchain, options, chain, entry, acceptCriteria, winnerCriteria string
	v.Proposal = NewEmptyProposalEntry()
	v.EligibleList = NewEligibleList()

	err := row.Scan(
		&vi,
		&sigKey,
		&sig,
		&v.Proposal.Proposal.Title,
		&v.Proposal.Proposal.Text,
		&v.Proposal.Proposal.ExternalRef.Href,
		&v.Proposal.Proposal.ExternalRef.Hash.Value, // External Hash
		&v.Proposal.Proposal.ExternalRef.Hash.Algo,
		&v.Proposal.Vote.PhasesBlockHeights.CommitStart,
		&v.Proposal.Vote.PhasesBlockHeights.CommitEnd,
		&v.Proposal.Vote.PhasesBlockHeights.RevealStart,
		&v.Proposal.Vote.PhasesBlockHeights.RevealEnd,
		&egchain,
		&v.Proposal.Vote.VoteType,
		&options,
		&v.Proposal.Vote.Config.AllowAbstention,
		&v.Proposal.Vote.Config.ComputeResultsAgainst,
		&v.Proposal.Vote.Config.MinOptions,
		&v.Proposal.Vote.Config.MaxOptions,
		&acceptCriteria,
		&winnerCriteria,
		&chain,
		&entry,
		&v.Proposal.BlockHeight,
		&v.Proposal.ProtocolVersion,
	)
	if err != nil {
		return nil, err
	}

	// Populate all the fields that have to be converted
	v.Proposal.VoteInitiator, _ = primitives.HexToHash(vi)

	key, _ := hex.DecodeString(sigKey)
	copy(v.Proposal.InitiatorKey[:], key[:])

	sigBytes, _ := hex.DecodeString(sig)
	v.Proposal.InitiatorSignature.SetSignature(sigBytes)

	egChainBytes, _ := hex.DecodeString(egchain)
	v.Proposal.Vote.EligibleVotersChainID.SetBytes(egChainBytes)

	v.Proposal.Vote.Config.Options = SplitString(options, ",")

	json.Unmarshal([]byte(acceptCriteria), &v.Proposal.Vote.Config.AcceptanceCriteria)
	json.Unmarshal([]byte(winnerCriteria), &v.Proposal.Vote.Config.WinnerCriteria)

	v.Proposal.ProposalChain, _ = primitives.HexToHash(chain)

	entryHashBytes, _ := hex.DecodeString(entry)
	v.Proposal.EntryHash.SetBytes(entryHashBytes)

	return v, nil
}

func (v Vote) SelectRows() string {
	return `vote_initiator,
			signing_key,
			signature,
			title,
			description,
			external_href,
			external_hash,
			external_hash_algo,
			commit_start,
			commit_stop,
			reveal_start,
			reveal_stop,
			eligible_voter_chain,
			vote_type,
			vote_options,
			vote_allow_abstain,
			vote_compute_results_against,
			vote_min_options,
			vote_max_options,
			vote_accept_criteria,
			vote_winner_criteria,
			chain_id,
			entry_hash,
			block_height,
			protocol_version`
}

func (v *Vote) RowValuePointers() []interface{} {
	data, _ := v.Proposal.InitiatorSignature.MarshalBinary()
	ac, _ := json.Marshal(v.Proposal.Vote.Config.AcceptanceCriteria)
	wc, _ := json.Marshal(v.Proposal.Vote.Config.WinnerCriteria)
	vi, sigKey, sig, exrefHash, egchain, options, chain, eHash, acceptCriteria, winnerCriteria :=
		v.Proposal.VoteInitiator.String(), // Vote Initiator
		v.Proposal.InitiatorKey.String(), // SigKey
		hex.EncodeToString(data), // Signature
		v.Proposal.Proposal.ExternalRef.Hash.Value, // External Hash
		v.Proposal.Vote.EligibleVotersChainID.String(), // Eligible Voter Chain
		strings.Join(v.Proposal.Vote.Config.Options, ","), // Vote Options
		v.Proposal.ProposalChain.String(),
		v.Proposal.EntryHash.String(),
		string(ac),
		string(wc)

	return []interface{}{
		&vi,
		&sigKey,
		&sig,
		&v.Proposal.Proposal.Title,
		&v.Proposal.Proposal.Text,
		&v.Proposal.Proposal.ExternalRef.Href,
		&exrefHash,
		&v.Proposal.Proposal.ExternalRef.Hash.Algo,
		&v.Proposal.Vote.PhasesBlockHeights.CommitStart,
		&v.Proposal.Vote.PhasesBlockHeights.CommitEnd,
		&v.Proposal.Vote.PhasesBlockHeights.RevealStart,
		&v.Proposal.Vote.PhasesBlockHeights.RevealEnd,
		&egchain,
		&v.Proposal.Vote.VoteType,
		&options,
		&v.Proposal.Vote.Config.AllowAbstention,
		&v.Proposal.Vote.Config.ComputeResultsAgainst,
		&v.Proposal.Vote.Config.MinOptions,
		&v.Proposal.Vote.Config.MaxOptions,
		&acceptCriteria,
		&winnerCriteria,
		&chain,
		&eHash,
		&v.Proposal.BlockHeight,
		&v.Proposal.ProtocolVersion}
}

// Commit

func (v *VoteCommit) New() ISQLObject {
	return NewVoteCommit()
}

func (v *VoteCommit) Table() string {
	return "commits"
}

func (v *VoteCommit) InsertFunction() string {
	return "insert_commit"
}

func (v *VoteCommit) ScanRow(row SQLRowWithScan) (*VoteCommit, error) {
	var id, sigKey, sig, chain, eHash string

	err := row.Scan(
		&id,
		&sigKey,
		&sig,
		&v.Content.Commitment,
		&chain,
		&eHash,
		&v.BlockHeight,
	)
	if err != nil {
		return nil, err
	}

	v.VoterID, _ = primitives.HexToHash(id)

	sigBytes, _ := hex.DecodeString(sig)
	v.Signature.UnmarshalBinary(sigBytes)

	v.VoteChain, _ = primitives.HexToHash(chain)
	v.EntryHash, _ = primitives.HexToHash(eHash)

	return v, nil
}

func (v VoteCommit) SelectRows() string {
	return `voter_id,
			signing_key,
			signature,
			commitment,
			vote_chain,
			entry_hash,
			block_height`
}

func (v *VoteCommit) RowValuePointers() []interface{} {
	data, _ := v.Signature.MarshalBinary()
	id, sigKey, sig, chain, ehash :=
		v.VoterID.String(), // Vote Initiator
		v.VoterKey.String(), // SigKey
		hex.EncodeToString(data), // Signature
		v.VoteChain.String(),
		v.EntryHash.String()

	return []interface{}{
		&id,
		&sigKey,
		&sig,
		&v.Content.Commitment,
		&chain,
		&ehash,
		&v.BlockHeight,
	}
}

// Reveal

func (v *VoteReveal) New() ISQLObject {
	return NewVoteReveal()
}

func (v *VoteReveal) Table() string {
	return "reveals"
}

func (v *VoteReveal) InsertFunction() string {
	return "insert_reveal"
}

func (v *VoteReveal) ScanRow(row SQLRowWithScan) (*VoteReveal, error) {
	var id, vote, chain, eHash string

	err := row.Scan(
		&id,
		&vote,
		&v.Content.Secret,
		&v.Content.HmacAlgo,
		&chain,
		&eHash,
		&v.BlockHeight,
	)
	if err != nil {
		return nil, err
	}

	v.Content.VoteOptions = SplitString(vote, ",")

	v.VoterID, _ = primitives.HexToHash(id)
	v.VoteChain, _ = primitives.HexToHash(chain)
	v.EntryHash, _ = primitives.HexToHash(eHash)

	return v, nil
}

func (v VoteReveal) SelectRows() string {
	return `voter_id,
			vote,
			secret,
			hmac_algo,
			vote_chain,
			entry_hash,
			block_height`
}

func (v *VoteReveal) RowValuePointers() []interface{} {
	id, vote, chain, ehash :=
		v.VoterID.String(), // Vote Initiator
		strings.Join(v.Content.VoteOptions, ","), // SigKey
		v.VoteChain.String(),
		v.EntryHash.String()

	return []interface{}{
		&id,
		&vote,
		&v.Content.Secret,
		&v.Content.HmacAlgo,
		&chain,
		&ehash,
		&v.BlockHeight,
	}
}

// EligibleList

func (v *EligibleList) New() ISQLObject {
	return NewEligibleList()
}

func (v *EligibleList) Table() string {
	return "eligible_list"
}

func (v *EligibleList) InsertFunction() string {
	return "insert_eligible_list"
}

func (v *EligibleList) ScanRow(row SQLRowWithScan) (*EligibleList, error) {
	var id, vi, nonce, key, sig string

	err := row.Scan(
		&id,
		&vi,
		&nonce,
		&key,
		&sig,
	)
	if err != nil {
		return nil, err
	}

	// TODO: Fill in

	return v, nil
}

func (v EligibleList) SelectRows() string {
	return `chain_id,
			vote_initiator,
			nonce,
			initiator_key,
			initiator_signature`
}

func (v *EligibleList) RowValuePointers() []interface{} {
	id, vi, nonce, key, sig :=
		v.ChainID.String(), // Chain_id
		v.EligibilityHeader.VoteInitiator.String(),
		v.EligibilityHeader.Nonce.String(),
		v.EligibilityHeader.InitiatorKey.String(),
		hex.EncodeToString(v.EligibilityHeader.InitiatorSignature.Bytes())

	return []interface{}{
		&id,
		&vi,
		&nonce,
		&key,
		&sig,
	}
}

// EligibleList

func (v *EligibleVoter) New() ISQLObject {
	return NewEligibleVoter()
}

func (v *EligibleVoter) Table() string {
	return "eligible_voters"
}

func (v *EligibleVoter) InsertFunction() string {
	return "insert_eligible_voter"
}

func (v *EligibleVoter) ScanRow(row SQLRowWithScan) (*EligibleVoter, error) {
	var id, list, ehash, keys string

	err := row.Scan(
		&id,
		&list,
		&v.VoteWeight,
		&ehash,
		&v.BlockHeight,
		&keys,
	)
	if err != nil {
		return nil, err
	}

	idBytes, _ := hex.DecodeString(id)
	v.VoterID.SetBytes(idBytes)

	listBytes, _ := hex.DecodeString(list)
	v.EligibleList.SetBytes(listBytes)

	entryBytes, _ := hex.DecodeString(ehash)
	v.EntryHash.SetBytes(entryBytes)

	v.SigningKeys = SplitString(keys, ",")

	return v, nil
}

func (v EligibleVoter) SelectRows() string {
	return `voter_id,
			eligible_list,
			weight,
			entry_hash,
			block_height,
			signing_keys`
}

func (v *EligibleVoter) RowValuePointers() []interface{} {
	id, list, ehash, keys :=
		v.VoterID.String(), // Chain_id
		v.EligibleList.String(),
		v.EntryHash.String(),
		strings.Join(v.SigningKeys, ",")

	return []interface{}{
		&id,
		&list,
		&v.VoteWeight,
		&ehash,
		&v.BlockHeight,
		&keys,
	}
}

// Results
func (v *VoteStats) New() ISQLObject {
	return NewVoteStats()
}

func (v *VoteStats) Table() string {
	return "results"
}

func (v *VoteStats) InsertFunction() string {
	return "insert_results"
}

func (v *VoteStats) ScanRow(row SQLRowWithScan) (*VoteStats, error) {
	var optJson, winJosn string

	err := row.Scan(
		&v.VoteChain,
		&v.Valid,
		&v.CompleteStats.Count,
		&v.CompleteStats.Weight,
		&v.VotedStats.Count,
		&v.VotedStats.Weight,
		&v.AbstainedStats.Count,
		&v.AbstainedStats.Weight,
		&v.Turnout.UnweightedTurnout,
		&v.Turnout.WeightedTurnout,
		&v.Support.CountDenominator,
		&v.Support.WeightDenominator,
		&optJson,
		&winJosn,
	)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(optJson), &v.OptionStats)
	json.Unmarshal([]byte(winJosn), &v.WeightedWinners)

	return v, nil
}

func (v VoteStats) SelectRows() string {
	return `vote_chain,
			valid_vote,
			complete_count,
			complete_weight,
			voted_count,
			voted_weight,
			abstained_count,
			abstained_weight,
			turnout_unweighted,
			turnout_weighted,
			support_unweighted,
			support_weighted,
			option_stats,
			winner_stats`
}

func (v *VoteStats) RowValuePointers() []interface{} {
	optBytes, _ := json.Marshal(&v.OptionStats)
	winBytes, _ := json.Marshal(&v.WeightedWinners)

	optJson, winJson := string(optBytes), string(winBytes)

	return []interface{}{
		&v.VoteChain,
		&v.Valid,
		&v.CompleteStats.Count,
		&v.CompleteStats.Weight,
		&v.VotedStats.Count,
		&v.VotedStats.Weight,
		&v.AbstainedStats.Count,
		&v.AbstainedStats.Weight,
		&v.Turnout.UnweightedTurnout,
		&v.Turnout.WeightedTurnout,
		&v.Support.CountDenominator,
		&v.Support.WeightDenominator,
		&optJson,
		&winJson,
	}
}
