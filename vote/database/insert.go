package database

import (
	"fmt"

	"github.com/Emyrk/go-factom-vote/vote/common"
)

func (db *SQLDatabase) InsertVote(v *common.Vote) error {
	query := fmt.Sprintf(`SELECT %s(%s)`, v.InsertFunction(), common.InsertQueryParams(v))
	fmt.Println(query)
	_, err := db.DB.Exec(query)
	return err
}
