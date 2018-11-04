CREATE OR REPLACE FUNCTION insert_commit(
  param_voter_id char(64),
  param_signing_key char(64),
  param_signature varchar,
  param_commitment varchar,
  param_vote_chain CHAR(64),
  param_entry_hash CHAR(64),
  param_block_height INTEGER)
  RETURNS INTEGER AS $$
BEGIN

  IF exists(SELECT voter_id, vote_chain FROM commits WHERE commits.voter_id = param_voter_id AND commits.vote_chain = param_vote_chain)
  THEN
    -- Data already exists in the table
    RETURN 0;
  ELSE
    -- Insert data into table
    INSERT INTO commits(voter_id,
                          signing_key,
                          signature,
                          commitment,
                          vote_chain,
                          entry_hash,
                          block_height)
    VALUES(param_voter_id,
      param_signing_key,
      param_signature,
      param_commitment,
      param_vote_chain,
      param_entry_hash,
      param_block_height);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql
