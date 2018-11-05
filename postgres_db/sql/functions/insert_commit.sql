CREATE OR REPLACE FUNCTION insert_commit(
  param_voter_id char(64),
  param_signing_key char(64),
  param_signature varchar,
  param_commitment varchar,
  param_vote_chain CHAR(64),
  param_entry_hash CHAR(64),
  param_block_height INTEGER)
  RETURNS INTEGER AS $$
  DECLARE
    com_start INTEGER;
    com_stop INTEGER;
    elig_chain CHAR(64);
BEGIN

  IF exists(SELECT vote_chain, voter_id, commitment FROM repeated_commits WHERE
    repeated_commits.vote_chain = param_vote_chain AND repeated_commits.voter_id = param_voter_id AND repeated_commits.commitment = param_commitment)
  THEN
    -- This is a replay
    RETURN 0;
  ELSE
    -- Need to determine the eligible list to use
    SELECT eligible_voter_chain INTO elig_chain FROM proposals WHERE chain_id = param_vote_chain;

    -- First check if the key is valid for the voter
    IF NOT exists(
      SELECT signing_keys FROM eligible_voters WHERE voter_id = param_voter_id
                                                     AND eligible_list = elig_chain
                                                     AND signing_keys LIKE concat('%', param_signing_key, '%')
    ) THEN
      -- This signing_key is not in the list of valid keys for the voter
      RETURN -2;
    END IF;

    -- Check if we are within the commitment phase
    SELECT commit_start, commit_stop INTO com_start, com_stop FROM proposals WHERE chain_id = param_vote_chain;
    IF param_block_height > com_stop OR param_block_height < com_start
      THEN
      -- Outside range of commitment phase
      RETURN -3;
    END IF;

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
      param_block_height)
    ON CONFLICT (voter_id, vote_chain) DO UPDATE
      SET
        signing_key = param_signing_key,
        signature = param_signature,
        commitment = param_commitment,
        entry_hash = param_entry_hash,
        block_height = param_block_height
    ;

    INSERT INTO repeated_commits(vote_chain, voter_id, commitment, block_height, entry_hash)
    VALUES (param_vote_chain, param_voter_id, param_commitment, param_block_height, param_entry_hash);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql
