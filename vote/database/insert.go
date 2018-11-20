package database

import (
	"fmt"

	"database/sql"

	"encoding/hex"

	"github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (db *SQLDatabase) InsertGenericTX(o common.ISQLObject, tx *sql.Tx) error {
	query := fmt.Sprintf(`SELECT %s(%s)`, o.InsertFunction(), common.InsertQueryParams(o))
	_, err := tx.Exec(query)
	return err
}

func (db *SQLDatabase) InsertAndQueryGeneric(o common.ISQLObject) (int, error) {
	query := fmt.Sprintf(`SELECT %s(%s)`, o.InsertFunction(), common.InsertQueryParams(o))
	row := db.DB.QueryRow(query)
	var i int
	err := row.Scan(&i)
	return i, err
}

func (db *SQLDatabase) InsertGeneric(o common.ISQLObject) error {
	query := fmt.Sprintf(`SELECT %s(%s)`, o.InsertFunction(), common.InsertQueryParams(o))
	_, err := db.DB.Exec(query)
	return err
}

func (db *SQLDatabase) SetRegistered(vote interfaces.IHash, registered bool) error {
	query := `UPDATE proposals SET registered = $2 WHERE chain_id = $1;`
	_, err := db.DB.Exec(query, vote.String(), registered)
	return err
}

func (db *SQLDatabase) InsertSubmittedHash(hash [32]byte, tx *sql.Tx) error {
	query := `INSERT INTO eligible_submitted(repeat_hash) VALUES ($1)`
	_, err := tx.Exec(query, hex.EncodeToString(hash[:]))
	return err
}

func (db *SQLDatabase) InsertCompleted(completed int) error {
	query := "INSERT INTO completed(block_height) VALUES($1)"
	_, err := db.DB.Exec(query, completed)
	return err
}
