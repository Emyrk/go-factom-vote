create table proposals
(
	vote_initiator char(64),
	signing_key char(64),
	signature char(128),
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
	vote_type varchar,
	vote_options varchar,
	vote_allow_abstain boolean,
	vote_compute_results_against varchar,
	vote_min_options integer,
	vote_max_options integer,
	chain_id char(64) not null
		constraint proposals_chain_id_pk
		primary key
)
;

