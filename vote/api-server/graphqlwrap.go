package apiserver

import (
	"fmt"

	"strings"

	"database/sql"

	"github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/Emyrk/go-factom-vote/vote/database"
)

var noextra []interface{}

// Wrapper for the sql db to have fetch functions that will be in the format for graphql
type GraphQLSQLDB struct {
	*database.SQLDatabase
}

var voterow = "vote_initiator, signing_key, signature, title, description, external_href, external_hash, external_hash_algo, commit_start, commit_stop, reveal_start, reveal_stop, eligible_voter_chain, vote_type, vote_options, vote_allow_abstain, vote_compute_results_against, vote_min_options, vote_max_options, vote_accept_criteria, vote_winner_criteria, chain_id, entry_hash, block_height, registered, complete"

var eligibleListRow = "chain_id, vote_initiator, nonce, initiator_key, initiator_signature"
var eligibleVoterRow = "voter_id, eligible_list, weight, entry_hash, block_height, signing_keys"
var commitRow = `voter_id, vote_chain, signing_key, signature, commitment, entry_hash, block_height`
var revealRow = `voter_id, vote_chain, vote, secret, hmac_algo, entry_hash, block_height`

func (g *GraphQLSQLDB) FetchVote(chainid string) (*Vote, error) {
	query := fmt.Sprintf(`SELECT %s FROM proposals WHERE chain_id = $1`, voterow)
	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("vote not found")
	}

	v := new(Vote)
	err = scanVote(rows, v, noextra)

	if err != nil {
		return nil, err
	}
	return v, nil
}

func (g *GraphQLSQLDB) FetchAllVoteStats(valid bool, offset int, limit int) (*VoteResultList, error) {
	r := new(common.VoteStats)
	where := ""
	var args []interface{}
	if valid {
		where = "WHERE valid_vote = $1 "
		args = append(args, valid)
	}
	query := fmt.Sprintf("SELECT %s, count(*) OVER() AS full_count FROM results %s", r.SelectRows(), where)
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := g.SQLDatabase.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []common.VoteStats
	count := new(int)
	for rows.Next() {
		v := new(common.VoteStats)
		err = scanVoteResults(rows, v, []interface{}{count})

		if err != nil {
			return nil, err
		}
		results = append(results, *v)
	}

	container := new(VoteResultList)
	container.Votes = results
	container.Info.TotalCount = *count
	container.Info.Offset = offset
	container.Info.Limit = limit

	return container, nil
}

func (g *GraphQLSQLDB) FetchVoteStats(chainid string) (*common.VoteStats, error) {
	r := new(common.VoteStats)
	query := fmt.Sprintf(`SELECT %s FROM results WHERE vote_chain = $1`, r.SelectRows())
	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("vote results not found")
	}

	r, err = r.ScanRow(rows)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (g *GraphQLSQLDB) FetchEligibleList(chainid string) (*EligibleList, error) {
	query := fmt.Sprintf(`SELECT %s FROM eligible_list WHERE chain_id = $1`, eligibleListRow)
	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("list not found")
	}

	e := new(EligibleList)
	err = rows.Scan(
		&e.Admin.ChainID,
		&e.Admin.Initiator,
		&e.Admin.Nonce,
		&e.Admin.SigningKey,
		&e.Admin.Signature,
	)

	if err != nil {
		return nil, err
	}
	return e, nil
}

func (g *GraphQLSQLDB) FetchEligibleVoters(chainid string, limit, offset int) (*EligibleVoterContainer, error) {
	query := fmt.Sprintf("SELECT %s, count(*) OVER() AS full_count FROM eligible_voters WHERE eligible_list = $1", eligibleVoterRow)
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []EligibleVoter
	count := new(int)
	for rows.Next() {
		v := new(EligibleVoter)
		err = scanEligibleVoter(rows, v, []interface{}{count})

		if err != nil {
			return nil, err
		}
		votes = append(votes, *v)
	}

	container := new(EligibleVoterContainer)
	container.EligibleVoters = votes
	container.Info.TotalCount = *count
	container.Info.Limit = limit
	container.Info.Offset = offset

	return container, nil
}

func scanEligibleVoter(rows *sql.Rows, v *EligibleVoter, extra []interface{}) error {
	var keys string
	var arr = []interface{}{
		&v.VoterID,
		&v.EligibleList,
		&v.VoteWeight,
		&v.EntryHash,
		&v.BlockHeight,
		&keys,
	}

	arr = append(arr, extra...)
	err := rows.Scan(
		arr...,
	)
	v.SigningKeys = strings.Split(keys, ",")
	return err
}

func scanVoteResults(rows *sql.Rows, v *common.VoteStats, extra []interface{}) error {
	var arr = v.RowValuePointers()

	arr = append(arr, extra...)
	err := rows.Scan(
		arr...,
	)
	return err
}

func (g *GraphQLSQLDB) FetchAllVotes(registered, active bool, limit, offset int) (*VoteList, error) {
	query := fmt.Sprintf(`SELECT %s, count(*) OVER() AS full_count FROM proposals`, voterow)

	where := ""
	if registered || active {
		where = " WHERE "
	}
	if registered {
		where += " registered = TRUE "
	}
	if registered && active {
		where += " AND "
	}
	if active {
		where += " reveal_stop > (SELECT max(block_height) FROM completed)"
	}
	query += where

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := g.SQLDatabase.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []Vote
	count := new(int)
	for rows.Next() {
		v := new(Vote)
		err = scanVote(rows, v, []interface{}{count})

		if err != nil {
			return nil, err
		}
		votes = append(votes, *v)
	}

	container := new(VoteList)
	container.Votes = votes
	container.Info.TotalCount = *count
	container.Info.Offset = offset
	container.Info.Limit = limit

	return container, nil
}

func scanVote(rows *sql.Rows, v *Vote, extra []interface{}) error {
	var options string
	var arr = []interface{}{
		// Admin Section
		&v.Admin.VoteInitator,
		&v.Admin.SigningKey,
		&v.Admin.Signature,
		&v.Admin.VoteInfo.Title,
		&v.Admin.VoteInfo.Text,
		&v.Admin.VoteInfo.ExternalRef.Href,
		&v.Admin.VoteInfo.ExternalRef.Hash.Value,
		&v.Admin.VoteInfo.ExternalRef.Hash.Algo,

		// Definition
		&v.Definition.PhasesBlockHeights.CommitStart,
		&v.Definition.PhasesBlockHeights.CommitStop,
		&v.Definition.PhasesBlockHeights.RevealStart,
		&v.Definition.PhasesBlockHeights.RevealStop,
		&v.Definition.EligibleVoterChain,

		// Config
		&v.Definition.VoteType,
		&options, // ->&v.Definition.Config.Options,
		&v.Definition.Config.AllowAbstention,
		&v.Definition.Config.ComputeResultsAgainst,
		&v.Definition.Config.MinOptions,
		&v.Definition.Config.MaxOptions,
		&v.Definition.Config.AcceptanceCriteria,
		&v.Definition.Config.WinnerCriteria,

		// Back to Admin
		&v.Chainid,
		&v.Admin.AdminEntryHash,
		&v.Admin.AdminBlockHeight,
		&v.Admin.Registered,
		&v.Admin.Complete,
	}

	arr = append(arr, extra...)
	err := rows.Scan(
		arr...,
	)
	v.Definition.Config.Options = strings.Split(options, ",")
	return err
}

func (g *GraphQLSQLDB) FetchAllCommits(chainid string, limit, offset int) (*VoteCommitContainer, error) {
	query := fmt.Sprintf(`SELECT %s, count(*) OVER() AS full_count FROM commits WHERE vote_chain = $1`, commitRow)

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []VoteCommit
	count := new(int)
	for rows.Next() {
		v := new(VoteCommit)
		err = scanCommit(rows, v, []interface{}{count})

		if err != nil {
			return nil, err
		}
		commits = append(commits, *v)
	}

	container := new(VoteCommitContainer)
	container.Commits = commits
	container.Info.TotalCount = *count
	container.Info.Offset = offset
	container.Info.Limit = limit

	return container, nil
}

func (g *GraphQLSQLDB) FetchCommit(voterID, voteChain string) (*VoteCommit, error) {
	query := fmt.Sprintf(`SELECT %s FROM commits WHERE voter_id = $1 AND vote_chain = $2`, commitRow)
	rows, err := g.SQLDatabase.DB.Query(query, voterID, voteChain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("list not found")
	}

	c := new(VoteCommit)
	err = scanCommit(rows, c, noextra)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func scanCommit(rows *sql.Rows, c *VoteCommit, extra []interface{}) error {
	var arr = []interface{}{
		&c.VoterID,
		&c.VoteChain,
		&c.SigningKey,
		&c.Signature,
		&c.Commitment,
		&c.EntryHash,
		&c.BlockHeight,
	}

	arr = append(arr, extra...)
	err := rows.Scan(
		arr...,
	)
	return err
}

func (g *GraphQLSQLDB) FetchAllReveals(chainid string, limit, offset int) (*VoteRevealContainer, error) {
	query := fmt.Sprintf(`SELECT %s, count(*) OVER() AS full_count FROM reveals WHERE vote_chain = $1`, revealRow)

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := g.SQLDatabase.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reveals []VoteReveal
	count := new(int)
	for rows.Next() {
		v := new(VoteReveal)
		err = scanReveal(rows, v, []interface{}{count})

		if err != nil {
			return nil, err
		}
		reveals = append(reveals, *v)
	}

	container := new(VoteRevealContainer)
	container.Reveals = reveals
	container.Info.TotalCount = *count
	container.Info.Offset = offset
	container.Info.Limit = limit

	return container, nil
}

func (g *GraphQLSQLDB) FetchReveal(voterID, voteChain string) (*VoteReveal, error) {
	query := fmt.Sprintf(`SELECT %s FROM reveals WHERE voter_id = $1 AND vote_chain = $2`, revealRow)
	rows, err := g.SQLDatabase.DB.Query(query, voterID, voteChain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("list not found")
	}

	r := new(VoteReveal)
	err = scanReveal(rows, r, noextra)

	if err != nil {
		return nil, err
	}
	return r, nil
}

func scanReveal(rows *sql.Rows, r *VoteReveal, extra []interface{}) error {
	var vote string
	var arr = []interface{}{
		&r.VoterID,
		&r.VoteChain,
		&vote,
		&r.Secret,
		&r.HmacAlgo,
		&r.EntryHash,
		&r.BlockHeight,
	}

	arr = append(arr, extra...)
	err := rows.Scan(
		arr...,
	)

	r.Vote = strings.Split(vote, ",")

	return err
}
