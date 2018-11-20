CREATE OR REPLACE FUNCTION insert_eligible_voter(
  param_voter_id char(64),
  param_eligible_list char(64),
  param_weight INTEGER,
  param_entry_hash CHAR(64),
  param_block_height INTEGER,
  param_signing_keys VARCHAR)
  RETURNS INTEGER AS $$
BEGIN

  IF exists(SELECT voter_id, eligible_list, entry_hash FROM eligible_voters WHERE
    eligible_voters.entry_hash = param_entry_hash AND
    eligible_voters.voter_id = param_voter_id AND
    eligible_voters.eligible_list = param_eligible_list)
  THEN
    -- This is a replay
    RETURN 0;
  END IF;

  -- Insert data into table
  INSERT INTO eligible_voters(voter_id,
                              eligible_list,
                              weight,
                              entry_hash,
                              block_height,
                              signing_keys)
  VALUES(param_voter_id,
         param_eligible_list,
         param_weight,
         param_entry_hash,
         param_block_height,
         param_signing_keys);
--     ON CONFLICT (voter_id, eligible_list) DO UPDATE
--     -- Update Weight
--     SET weight = param_weight,
--       entry_hash = param_entry_hash,
--       block_height = param_block_height;
    RETURN 1;
END;
$$ LANGUAGE plpgsql