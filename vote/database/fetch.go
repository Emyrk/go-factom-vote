package database

import (
	"database/sql"

	"fmt"

	"github.com/Emyrk/go-factom-vote/vote/common"
)

func (s *SQLDatabase) FetchHighestDBInserted() int {
	highest := -1
	row := s.QueryRow(`SELECT MAX(block_height) FROM completed`)
	row.Scan(&highest)
	return highest // Highest will be -1 in the case of no rows found, which is fine
}

func (s *SQLDatabase) IsRepeatedEntryExists(hash string) (bool, error) {
	query := `SELECT repeat_hash FROM eligible_submitted WHERE repeat_hash = $1`
	return exists(s.DB.Query(query, hash))
}

func (s *SQLDatabase) IsVoteExist(voteId string) (bool, error) {
	query := `SELECT chain_id FROM proposals WHERE chain_id = $1`
	return exists(s.DB.Query(query, voteId))
}

func (s *SQLDatabase) IsEligibleListExist(chainId string) (bool, error) {
	query := `SELECT chain_id FROM eligible_list WHERE chain_id = $1`
	return exists(s.DB.Query(query, chainId))
}

func (s *SQLDatabase) IsEligibleListExistWithKey(chainId string) (bool, string, error) {
	query := `SELECT (chain_id, initiator_key) FROM eligible_list WHERE chain_id = $1`
	rows, err := s.DB.Query(query, chainId)
	if err != nil {
		return false, "", err
	}
	defer rows.Close()

	if rows.Next() {
		var chain, key string
		err = rows.Scan(&chain, &key)
		if err != nil {
			return false, "", err
		}

		return true, key, nil
	}
	return false, "", nil
}

func exists(rows *sql.Rows, err error) (bool, error) {
	if !rows.Next() {
		return false, nil
	}
	rows.Close()

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

type PartialCommit struct {
	VoterID    string
	SigningKey string
	Commitment string
	VoteChain  string
}

func (s *SQLDatabase) FetchCommitForReveal(reveal common.VoteReveal) (*PartialCommit, error) {
	pc := new(PartialCommit)

	query := `SELECT voter_id, signing_key, commitment, vote_chain FROM commits WHERE 
				voter_id = $1 AND vote_chain = $2`
	row := s.DB.QueryRow(query, reveal.VoterID.String(), reveal.VoteChain.String())
	err := row.Scan(&pc.VoterID, &pc.SigningKey, &pc.Commitment, &pc.VoteChain)
	if err != nil {
		return nil, err
	}

	return pc, nil
}

func (s *SQLDatabase) FetchVote(chainid string) (*common.Vote, error) {
	v := new(common.Vote)
	var err error

	query := fmt.Sprintf("SELECT %s FROM %s WHERE chain_id = $1", v.SelectRows(), v.Table())
	row := s.DB.QueryRow(query, chainid)
	v, err = v.ScanRow(row)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *SQLDatabase) FetchCompleteVotes(height int) ([]*common.Vote, error) {
	v := new(common.Vote)
	var err error
	var votes []*common.Vote

	query := fmt.Sprintf("SELECT %s FROM %s WHERE reveal_stop = $1", v.SelectRows(), v.Table())
	rows, err := s.DB.Query(query, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := new(common.Vote)
		v, err = tmp.ScanRow(rows)
		if err != nil {
			return nil, err
		}

		votes = append(votes, tmp)
	}

	return votes, nil
}

func (s *SQLDatabase) FetchEligibleVoters(chainid string) ([]*common.EligibleVoter, error) {
	v := new(common.EligibleVoter)
	var err error

	query := fmt.Sprintf("SELECT %s FROM %s WHERE eligible_list = $1", v.SelectRows(), v.Table())
	rows, err := s.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr []*common.EligibleVoter
	for rows.Next() {
		n := new(common.EligibleVoter)
		n, err = n.ScanRow(rows)
		if err == nil {
			arr = append(arr, n)
		}
	}

	return arr, nil
}

func (s *SQLDatabase) FetchCommits(chainid string) ([]*common.VoteCommit, error) {
	v := new(common.VoteCommit)
	var err error

	query := fmt.Sprintf("SELECT %s FROM %s WHERE vote_chain = $1", v.SelectRows(), v.Table())
	rows, err := s.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr []*common.VoteCommit
	for rows.Next() {
		n := new(common.VoteCommit)
		n, err = n.ScanRow(rows)
		if err == nil {
			arr = append(arr, n)
		}
	}

	return arr, nil
}

func (s *SQLDatabase) FetchReveals(chainid string) ([]*common.VoteReveal, error) {
	v := new(common.VoteReveal)
	var err error

	query := fmt.Sprintf("SELECT %s FROM %s WHERE vote_chain = $1", v.SelectRows(), v.Table())
	rows, err := s.DB.Query(query, chainid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arr []*common.VoteReveal
	for rows.Next() {
		n := new(common.VoteReveal)
		n, err = n.ScanRow(rows)
		if err == nil {
			arr = append(arr, n)
		}
	}

	return arr, nil
}
