CREATE OR REPLACE FUNCTION insert_results(
  param_vote_chain CHAR(64),
  param_valid_vote BOOLEAN,
  param_complete_count FLOAT,
  param_complete_weight FLOAT,
  param_voted_count FLOAT,
  param_voted_weight integer,
  param_abstained_count FLOAT,
  param_abstained_weight FLOAT,
  param_turnout_unweighted FLOAT,
  param_turnout_weighted FLOAT,
  param_support_unweighted FLOAT,
  param_support_weighted FLOAT,
  param_option_stats VARCHAR,
  param_winner_stats VARCHAR)
  RETURNS INTEGER AS $$
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
$$ LANGUAGE plpgsql
