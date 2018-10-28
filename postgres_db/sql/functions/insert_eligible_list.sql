CREATE OR REPLACE FUNCTION insert_eligible_list(
  param_chain_id char(64),
  param_vote_initiator char(64),
  param_nonce VARCHAR,
  param_initiator_key VARCHAR,
  param_initiator_signature VARCHAR)
  RETURNS INTEGER AS $$
BEGIN
  IF exists(SELECT chain_id FROM eligible_list WHERE eligible_list.chain_id = param_chain_id)
  THEN
    -- Already exists
    RETURN 0;
  ELSE
    -- Insert data into table
    INSERT INTO eligible_list(chain_id,
                                vote_initiator,
                                nonce,
                                initiator_key,
                                initiator_signature)
    VALUES(param_chain_id,
          param_vote_initiator,
          param_nonce,
          param_initiator_key,
          param_initiator_signature);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql