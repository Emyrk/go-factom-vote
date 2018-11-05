CREATE OR REPLACE FUNCTION insert_reveal(
  param_voter_id char(64),
  param_vote VARCHAR,
  param_secret varchar,
  param_hmac_algo varchar,
  param_vote_chain CHAR(64),
  param_entry_hash CHAR(64),
  param_block_height INTEGER)
  RETURNS INTEGER AS $$
  DECLARE
    rev_start INTEGER;
    rev_stop INTEGER;
    elig_chain CHAR(64);
BEGIN

  IF exists(SELECT vote_chain, voter_id, vote FROM repeated_reveals WHERE
    repeated_reveals.vote_chain = param_vote_chain AND repeated_reveals.voter_id = param_voter_id AND repeated_reveals.vote = param_vote)
  THEN
    -- This is a replay
    RETURN 0;
  ELSE

    -- Check if we are within the commitment phase
    SELECT reveal_start, reveal_stop INTO rev_start, rev_stop FROM proposals WHERE chain_id = param_vote_chain;
    IF param_block_height > rev_stop OR param_block_height < rev_start
    THEN
      -- Outside range of reveal phase
      RETURN -3;
    END IF;

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

    INSERT INTO repeated_reveals(vote_chain, voter_id, vote, block_height, entry_hash)
    VALUES (param_vote_chain, param_voter_id, param_vote, param_block_height, param_entry_hash);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql
