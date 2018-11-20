TRUNCATE eligible_voters, commits, reveals, eligible_list, eligible_submitted, proposals,
repeated_commits, repeated_reveals, results;

DELETE FROM completed WHERE block_height > 49000;