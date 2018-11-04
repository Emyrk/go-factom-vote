CREATE OR REPLACE FUNCTION insert_reveal(
  param_voter_id char(64),
  param_vote char(64),
  param_secret varchar,
  param_hmac_algo varchar,
  param_vote_chain CHAR(64),
  param_entry_hash CHAR(64),
  param_block_height INTEGER)
  RETURNS INTEGER AS $$
BEGIN

  IF exists(SELECT voter_id, vote_chain FROM reveals WHERE reveals.voter_id = param_voter_id AND reveals.vote_chain = param_vote_chain)
  THEN
    -- Data already exists in the table
    RETURN 0;
  ELSE
    -- Insert data into table
    INSERT INTO reveals(voter_id,
                        vote,
                        secret,
                        hmac_algo,
                        vote_chain,
                        entry_hash,
                        block_height)
    VALUES(param_voter_id,
           param_vote,
           param_secret,
           param_hmac_algo,
           param_vote_chain,
           param_entry_hash,
          param_block_height);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql
