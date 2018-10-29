package database

import "database/sql"

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

func exists(rows *sql.Rows, err error) (bool, error) {
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}
