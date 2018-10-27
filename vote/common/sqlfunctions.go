package common

import (
	"encoding/hex"
	"strings"
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
	var vi, sigKey, sig, hash, egchain, options, chain string

	err := row.Scan(
		&vi,
		&sigKey,
		&sig,
		&v.Proposal.Proposal.Title,
		&v.Proposal.Proposal.Text,
		&v.Proposal.Proposal.ExternalRef.Href,
		&hash,
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
		&chain,
	)
	if err != nil {
		return nil, err
	}
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
			chain_id`
}

func (v *Vote) RowValuePointers() []interface{} {
	data, _ := v.Proposal.InitiatorSignature.MarshalBinary()
	vi, sigKey, sig, exrefHash, egchain, options, chain :=
		v.Proposal.VoteInitiator.String(), // Vote Initiator
		v.Proposal.InitiatorKey.String(), // SigKey
		hex.EncodeToString(data), // Signature
		v.Proposal.Proposal.ExternalRef.Hash.Value.String(), // External Hash
		v.Proposal.Vote.EligibleVotersChainID.String(), // Eligible Voter Chain
		strings.Join(v.Proposal.Vote.Config.Options, ","), // Vote Options
		v.Proposal.ProposalChain.String()

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
		&chain}
}
