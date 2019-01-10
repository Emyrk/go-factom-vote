CREATE OR REPLACE FUNCTION fetch_proposal_entries(
  param_vote_chain  CHAR(64)
)
  RETURNS TABLE (
    voter_id CHAR(64),
    commit CHAR(64),
    reveal CHAR(64)
  )
AS $$
DECLARE
  vote_eligible_list CHAR(64);
  block_height INTEGER;
BEGIN
  SELECT eligible_voter_chain, commit_start INTO vote_eligible_list, block_height FROM proposals WHERE chain_id = param_vote_chain;

  RETURN QUERY
    SELECT voters.voter_id, coms.entry_hash, revs.entry_hash
    FROM fetch_eligible_voters(vote_eligible_list, block_height) AS voters
      LEFT JOIN
      (SELECT commits.voter_id, commits.entry_hash FROM commits
      WHERE commits.vote_chain = param_vote_chain) AS coms
        ON coms.voter_id = voters.voter_id
      LEFT JOIN
      (SELECT reveals.voter_id, reveals.entry_hash FROM reveals
      WHERE reveals.vote_chain = param_vote_chain) AS revs
        ON revs.voter_id = voters.voter_id;
END;
$$ LANGUAGE plpgsql;