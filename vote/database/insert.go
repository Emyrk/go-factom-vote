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

func (db *SQLDatabase) InsertGeneric(o common.ISQLObject) error {
	query := fmt.Sprintf(`SELECT %s(%s)`, o.InsertFunction(), common.InsertQueryParams(o))
	_, err := db.DB.Exec(query)
	return err
}

func (db *SQLDatabase) SetRegistered(vote interfaces.IHash) error {
	query := `UPDATE proposals SET registered = True WHERE chain_id = $1;`
	_, err := db.DB.Exec(query, vote.String())
	return err
}

func (db *SQLDatabase) InsertSubmittedHash(hash [32]byte, tx *sql.Tx) error {
	query := `INSERT INTO eligible_submitted VALUES ($1)`
	_, err := tx.Exec(query, hex.EncodeToString(hash[:]))
	return err
}
