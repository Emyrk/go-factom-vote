package vote

import (
	"database/sql"

	"fmt"

	"strings"

	"encoding/hex"

	. "github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/interfaces"
)

// All vote modifications go through here
func (vw *VoteWatcher) AddNewVoteProposal(v *Vote) error {
	err := vw.SQLDB.InsertGeneric(v)
	return err
}

func (vw *VoteWatcher) AddReveal(r VoteReveal, height uint32) error {
	// Find commit
	partialCommit, err := vw.SQLDB.FetchCommitForReveal(r)
	if err != nil {
		return fmt.Errorf("(add:fetchCommit) %s", err.Error())
	}

	// Validate reveal against commit
	commitment, _ := hex.DecodeString(partialCommit.Commitment)
	secret, _ := hex.DecodeString(r.Content.Secret)

	if !CheckMAC(r.Content.HmacAlgo,
		[]byte(strings.Join(r.Content.VoteOptions, "")),
		commitment,
		secret) {
		return fmt.Errorf("reveal does not validate hmac against commit.")
	}

	err = vw.SQLDB.InsertGeneric(&r)
	if err != nil {
		return fmt.Errorf("(add:insert) %s", err.Error())
	}
	return nil
}

func (vw *VoteWatcher) AddCommit(c VoteCommit, height uint32) error {
	err := vw.SQLDB.InsertGeneric(&c)
	if err != nil {
		return err
	}
	return nil
}

func (vw *VoteWatcher) SetRegistered(chain interfaces.IHash, registered bool) error {
	return vw.SQLDB.SetRegistered(chain, registered)
}

func (vw *VoteWatcher) AddNewEligibleList(e *EligibleList, hash [32]byte) error {
	err := vw.SQLDB.InsertGeneric(e)
	if err != nil {
		return err
	}

	tx, err := vw.SQLDB.Begin()
	if err != nil {
		return err
	}

	for _, v := range e.EligibleVoters {
		err := vw.addVoter(&v, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = vw.SQLDB.InsertSubmittedHash(hash, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (vw *VoteWatcher) AddEligibleVoter(voter *EligibleVoterEntry, hash [32]byte) error {
	tx, err := vw.SQLDB.Begin()
	if err != nil {
		return err
	}

	for _, v := range voter.Content {
		err := vw.addVoter(&v, tx)
		if err != nil {
			return err
		}
	}

	err = vw.SQLDB.InsertSubmittedHash(hash, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (vw *VoteWatcher) addVoter(voter *EligibleVoter, tx *sql.Tx) error {
	// Must get all the voting keys for this voter
	id := &factom.Identity{}
	id.ChainID = voter.VoterID.String()
	keys, err := id.GetKeysAtHeight(int64(voter.BlockHeight))
	if err != nil {
		return err
	}

	for _, k := range keys {
		voter.SigningKeys = append(voter.SigningKeys, fmt.Sprintf("%x", k.Pub[:]))
	}

	return vw.SQLDB.InsertGenericTX(voter, tx)
}

// Retrieval based questions
func (vw *VoteWatcher) IsEligibleListExist(chainid string) (bool, error) {
	return vw.SQLDB.IsEligibleListExist(chainid)
}

func (vw *VoteWatcher) IsEligibleListExistWithKey(chainid string) (bool, string, error) {
	return vw.SQLDB.IsEligibleListExistWithKey(chainid)
}
