CREATE OR REPLACE FUNCTION fetch_eligible_voters(
  param_eligible_list  CHAR(64),
  param_block_height INTEGER
)
  RETURNS TABLE (
    voter_id CHAR(64),
    eligible_list CHAR(64),
    weight DOUBLE PRECISION,
    entry_hash CHAR(64),
    block_height INTEGER,
    signing_keys VARCHAR,
    full_count BIGINT
  )
AS $$
BEGIN
  RETURN QUERY SELECT eligible_voters.voter_id, eligible_voters.eligible_list, eligible_voters.weight,
                 eligible_voters.entry_hash, eligible_voters.block_height, eligible_voters.signing_keys, count(*) OVER() AS full_count
               FROM eligible_voters
                 RIGHT JOIN
                 (SELECT eligible_voters.voter_id, max(eligible_voters.block_height) AS block_height FROM eligible_voters WHERE
                   eligible_voters.eligible_list = param_eligible_list AND eligible_voters.block_height < param_block_height GROUP BY (eligible_voters.voter_id)) AS maximums
                   ON eligible_voters.voter_id = maximums.voter_id AND eligible_voters.block_height = maximums.block_height WHERE eligible_voters.eligible_list = param_eligible_list;
END;
$$ LANGUAGE plpgsql;