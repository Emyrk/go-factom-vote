create table completed
(
  block_height integer not null
    constraint cmpleted_pkey
    primary key
)
;

create table proposals
(
  vote_initiator char(64),
  signing_key char(64),
  signature varchar,
  title varchar,
  description varchar,
  external_href varchar,
  external_hash varchar,
  external_hash_algo varchar,
  commit_start integer,
  commit_stop integer,
  reveal_start integer,
  reveal_stop integer,
  eligible_voter_chain char(64),
  vote_type integer,
  vote_options varchar,
  vote_allow_abstain boolean,
  vote_compute_results_against varchar,
  vote_min_options integer,
  vote_max_options integer,
  chain_id char(64) not null
    constraint proposals_chain_id_pk
    primary key,
  registered boolean default false,
  entry_hash char(64),
  block_height integer,
  vote_accept_criteria varchar,
  vote_winner_criteria varchar,
  complete boolean default false,
  protocol_version integer default 0
)
;

comment on column proposals.vote_accept_criteria is 'Raw JSON'
;

comment on column proposals.vote_winner_criteria is 'Raw JSON'
;

create table commits
(
  voter_id char(64),
  signing_key char(64),
  signature varchar,
  commitment varchar,
  id serial not null
    constraint commits_id_pk
    primary key,
  vote_chain char(64),
  entry_hash char(64),
  block_height integer
)
;

create unique index commits_vote_index
  on commits (voter_id, vote_chain)
;

create index commits_vote_chain_index
  on commits (vote_chain)
;

create table reveals
(
  voter_id char(64),
  vote varchar,
  secret varchar,
  hmac_algo varchar,
  id serial not null
    constraint reveals_id_pk
    primary key,
  vote_chain char(64),
  entry_hash char(64),
  block_height integer
)
;

create unique index reveals_id_uindex
  on reveals (id)
;

create index reveals_vote_index
  on reveals (voter_id, vote_chain)
;

create table eligible_list
(
  chain_id char(64) not null
    constraint eligible_list_pkey
    primary key,
  vote_initiator char(64),
  nonce varchar,
  initiator_key varchar,
  initiator_signature varchar
)
;

create table eligible_voters
(
  voter_id char(64) not null,
  eligible_list char(64) not null,
  weight double precision,
  entry_hash char(64) not null,
  block_height integer,
  signing_keys varchar,
  id serial not null
    constraint eligible_voters_pk_2
    primary key
)
;

create index eligible_voters_eligiblelist_index
  on eligible_voters (eligible_list)
;

create index eligible_voters_block_height_index
  on eligible_voters (block_height)
;

create table eligible_submitted
(
  repeat_hash char(64) not null
    constraint eligible_submitted_pkey
    primary key
)
;

create table repeated_commits
(
  commitment varchar not null,
  voter_id char(64) not null,
  vote_chain char(64) not null,
  block_height integer,
  entry_hash char(64),
  constraint repeated_commits_vote_chain_voter_id_commitment_pk
  primary key (vote_chain, voter_id, commitment)
)
;

create table repeated_reveals
(
  vote varchar not null,
  vote_chain char(64) not null,
  block_height integer,
  entry_hash char(64),
  voter_id char(64) not null,
  constraint repeated_reveals_vote_chain_voter_id_vote_pk
  primary key (vote_chain, voter_id, vote)
)
;

create table results
(
  vote_chain char(64) not null
    constraint results_pkey
    primary key,
  valid_vote boolean,
  complete_count double precision,
  complete_weight double precision,
  voted_count double precision,
  voted_weight integer,
  abstained_count double precision,
  abstained_weight double precision,
  turnout_unweighted double precision,
  turnout_weighted double precision,
  support_unweighted double precision,
  support_weighted double precision,
  option_stats varchar,
  winner_stats varchar
)
;

create unique index results_vote_chain_uindex
  on results (vote_chain)
;

comment on table results is 'result of vote when complete (passed reveal phase)'
;

create function insert_commit(param_voter_id character, param_signing_key character, param_signature character varying, param_commitment character varying, param_vote_chain character, param_entry_hash character, param_block_height integer) returns integer
language plpgsql
as $$
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
$$
;

create function insert_eligible_list(param_chain_id character, param_vote_initiator character, param_nonce character varying, param_initiator_key character varying, param_initiator_signature character varying) returns integer
language plpgsql
as $$
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
$$
;

create function insert_reveal(param_voter_id character, param_vote character varying, param_secret character varying, param_hmac_algo character varying, param_vote_chain character, param_entry_hash character, param_block_height integer) returns integer
language plpgsql
as $$
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
$$
;

create function insert_results(param_vote_chain character, param_valid_vote boolean, param_complete_count double precision, param_complete_weight double precision, param_voted_count double precision, param_voted_weight integer, param_abstained_count double precision, param_abstained_weight double precision, param_turnout_unweighted double precision, param_turnout_weighted double precision, param_support_unweighted double precision, param_support_weighted double precision, param_option_stats character varying, param_winner_stats character varying) returns integer
language plpgsql
as $$
DECLARE
BEGIN

  IF exists(SELECT vote_chain FROM results WHERE
    results.vote_chain = param_vote_chain)
  THEN
    -- This is a repeat
    RETURN 0;
  ELSE
    -- Insert data into table
    INSERT INTO results(vote_chain,
                        valid_vote,
                        complete_count,
                        complete_weight,
                        voted_count,
                        voted_weight,
                        abstained_count,
                        abstained_weight,
                        turnout_unweighted,
                        turnout_weighted,
                        support_unweighted,
                        support_weighted,
                        option_stats,
                        winner_stats)
    VALUES(param_vote_chain,
      param_valid_vote,
      param_complete_count,
      param_complete_weight,
      param_voted_count,
      param_voted_weight,
      param_abstained_count,
      param_abstained_weight,
      param_turnout_unweighted,
      param_turnout_weighted,
      param_support_unweighted,
           param_support_weighted,
           param_option_stats,
           param_winner_stats);

    UPDATE proposals SET complete = True WHERE chain_id = param_vote_chain;
    RETURN 1;
  end if;
  RETURN -1;
END;
$$
;

create function insert_eligible_voter(param_voter_id character, param_eligible_list character, param_weight double precision, param_entry_hash character, param_block_height integer, param_signing_keys character varying) returns integer
language plpgsql
as $$
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
$$
;

create function insert_vote(param_vote_initiator character, param_signing_key character, param_signature character varying, param_title character varying, param_description character varying, param_external_href character varying, param_external_hash character varying, param_external_hash_algo character varying, param_commit_start integer, param_commit_stop integer, param_reveal_start integer, param_reveal_stop integer, param_eligible_voter_chain character, param_vote_type integer, param_vote_options character varying, param_vote_allow_abstain boolean, param_vote_compute_results_against character varying, param_vote_min_options integer, param_vote_max_options integer, param_vote_accept_criteria character varying, param_vote_winner_criteria character varying, param_chain_id character, param_entry_hash character, param_block_height integer, param_protocol_version integer) returns integer
language plpgsql
as $$
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
                          vote_accept_criteria,
                          vote_winner_criteria,
                          chain_id,
                          entry_hash,
                          block_height,
                          protocol_version,
                          registered)
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
      param_vote_accept_criteria,
      param_vote_winner_criteria,
           param_chain_id,
           param_entry_hash,
           param_block_height,
           param_protocol_version,
           FALSE);
    RETURN 1;
  end if;
  RETURN -1;
END;
$$
;

create function fetch_eligible_voters(param_eligible_list character, param_block_height integer) returns TABLE(voter_id character, eligible_list character, weight double precision, entry_hash character, block_height integer, signing_keys character varying, full_count bigint)
language plpgsql
as $$
BEGIN
  RETURN QUERY SELECT eligible_voters.voter_id, eligible_voters.eligible_list, eligible_voters.weight,
                 eligible_voters.entry_hash, eligible_voters.block_height, eligible_voters.signing_keys, count(*) OVER() AS full_count
               FROM eligible_voters
                 RIGHT JOIN
                 (SELECT eligible_voters.voter_id, max(eligible_voters.block_height) AS block_height FROM eligible_voters WHERE
                   eligible_voters.eligible_list = param_eligible_list AND eligible_voters.block_height < param_block_height GROUP BY (eligible_voters.voter_id)) AS maximums
                   ON eligible_voters.voter_id = maximums.voter_id AND eligible_voters.block_height = maximums.block_height WHERE eligible_voters.eligible_list = param_eligible_list;
END;
$$
;

create function fetch_proposal_entries(param_vote_chain character) returns TABLE(voter_id character, commit character, reveal character)
language plpgsql
as $$
DECLARE
  vote_eligible_list CHAR(64);
  block_height INTEGER;
BEGIN
  SELECT eligible_voter_chain, commit_start INTO vote_eligible_list, block_height FROM proposals WHERE chain_id = param_vote_chain;

  RAISE NOTICE '%', block_height;
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
$$
;

