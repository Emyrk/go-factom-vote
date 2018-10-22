package database

import (
	"fmt"

	"github.com/Emyrk/go-factom-vote/vote"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var _ = log.Errorf

func (db *SQLDatabase) InsertProposal(vote *vote.Vote) error {
	db.DB.Exec(fmt.Sprintf(`
	SELECT insert_vote`))

	return nil
}
