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
	registered boolean
)
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

create index commits_vote_index
	on commits (voter_id, vote_chain)
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
	weight integer,
	entry_hash char(64) not null,
	block_height integer,
	constraint eligible_voters_pk
	primary key (voter_id, eligible_list)
)
;

create table eligible_submitted
(
	repeat_hash char(64) not null
		constraint eligible_submitted_pkey
		primary key
)
;

create function insert_vote(param_vote_initiator character, param_signing_key character, param_signature character varying, param_title character varying, param_description character varying, param_external_href character varying, param_external_hash character varying, param_external_hash_algo character varying, param_commit_start integer, param_commit_stop integer, param_reveal_start integer, param_reveal_stop integer, param_eligible_voter_chain character, param_vote_type integer, param_vote_options character varying, param_vote_allow_abstain boolean, param_vote_compute_results_against character varying, param_vote_min_options integer, param_vote_max_options integer, param_chain_id character) returns integer
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
$$
;

create function insert_reveal(param_voter_id character, param_vote character, param_secret character varying, param_hmac_algo character varying, param_vote_chain character, param_entry_hash character, param_block_height integer) returns integer
language plpgsql
as $$
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
$$
;

create function insert_commit(param_voter_id character, param_signing_key character, param_signature character varying, param_commitment character varying, param_vote_chain character, param_entry_hash character, param_block_height integer) returns integer
language plpgsql
as $$
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
$$
;

create function insert_eligible_voter(param_voter_id character, param_eligible_list character, param_weight integer, param_entry_hash character, param_block_height integer) returns integer
language plpgsql
as $$
BEGIN
	IF exists(SELECT voter_id, eligible_list FROM eligible_voters WHERE eligible_voters.voter_id = param_voter_id AND eligible_voters.eligible_list = param_eligible_list) AND param_weight = 0
	THEN
		-- Removing an eligible voter
		DELETE FROM eligible_voters WHERE voter_id = param_voter_id AND eligible_list = param_eligible_list;
		RETURN 0;
	ELSE
		-- Insert data into table
		INSERT INTO eligible_voters(voter_id,
																eligible_list,
																weight,
																entry_hash,
																block_height)
		VALUES(param_voter_id,
					 param_eligible_list,
					 param_weight,
					 param_entry_hash,
					 param_block_height)
		ON CONFLICT (voter_id, eligible_list) DO UPDATE
			-- Update Weight
			SET weight = param_weight;
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

