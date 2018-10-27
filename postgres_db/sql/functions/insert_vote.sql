CREATE OR REPLACE FUNCTION insert_vote(
  param_vote_initiator char(64),
  param_signing_key char(64),
  param_signature varchar,
  param_title varchar,
  param_description varchar,
  param_external_href varchar,
  param_external_hash varchar,
  param_external_hash_algo varchar,
  param_commit_start integer,
  param_commit_stop integer,
  param_reveal_start integer,
  param_reveal_stop integer,
  param_eligible_voter_chain char(64),
  param_vote_type INTEGER,
  param_vote_options varchar,
  param_vote_allow_abstain boolean,
  param_vote_compute_results_against varchar,
  param_vote_min_options integer,
  param_vote_max_options integer,
  param_chain_id char(64))
  RETURNS INTEGER AS $$
BEGIN

  IF exists(SELECT chain_id FROM proposals WHERE proposals.chain_id = param_chain_id)
  THEN
    -- Data already exists in the table
    RETURN 0;
  ELSE
    -- Insert data into table
    INSERT INTO proposals(vote_initiator,
     signing_key,
     signature,
     title,
     description,
     external_href,
     external_hash,
     external_hash_algo,
     commit_start,
     commit_stop,
     reveal_start,
     reveal_stop,
     eligible_voter_chain,
     vote_type,
     vote_options,
     vote_allow_abstain,
     vote_compute_results_against,
     vote_min_options,
     vote_max_options,
     chain_id)
    VALUES(param_vote_initiator,
      param_signing_key,
      param_signature,
      param_title,
      param_description,
      param_external_href,
      param_external_hash,
      param_external_hash_algo,
      param_commit_start,
      param_commit_stop,
      param_reveal_start,
      param_reveal_stop,
      param_eligible_voter_chain,
      param_vote_type,
      param_vote_options,
      param_vote_allow_abstain,
      param_vote_compute_results_against,
      param_vote_min_options,
      param_vote_max_options,
      param_chain_id);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$ LANGUAGE plpgsql
