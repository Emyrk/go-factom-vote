package vote

import (
	"fmt"

	. "github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/FactomProject/factomd/common/interfaces"
)

// All vote modifications go through here
func (vw *VoteWatcher) AddNewVoteProposal(v *Vote) error {
	if vw.UseMemory {
		vw.VoteProposals[v.Proposal.ProposalChain.Fixed()] = v
		return nil
	}
	err := vw.SQLDB.InsertGeneric(v)
	return err
}

func (vw *VoteWatcher) AddReveal(v *Vote, r VoteReveal, height uint32) error {
	if vw.UseMemory {
		return v.AddReveal(r, height)
	}
	err := vw.SQLDB.InsertGeneric(&r)
	if err != nil {
		return err
	}
	return nil
}

func (vw *VoteWatcher) AddCommit(v *Vote, c VoteCommit, height uint32) error {
	if vw.UseMemory {
		return v.AddCommit(c, height)
	}
	err := vw.SQLDB.InsertGeneric(&c)
	if err != nil {
		return err
	}
	return nil
}

func (vw *VoteWatcher) SetRegistered(chain interfaces.IHash, registered bool) error {
	if vw.UseMemory {
		v, _ := vw.VoteProposals[chain.Fixed()]
		if v != nil {
			v.Registered = registered
		}
		return nil
	}

	return vw.SQLDB.SetRegistered(chain, registered)
}

func (vw *VoteWatcher) AddNewEligibleList(e *EligibleList, hash [32]byte) error {
	if vw.UseMemory {
		e.SubmittedEntries[hash] = true
		vw.EligibleLists[e.ChainID.Fixed()] = e
		return nil
	}

	err := vw.SQLDB.InsertGeneric(e)
	if err != nil {
		return err
	}

	tx, err := vw.SQLDB.Begin()
	if err != nil {
		return err
	}

	for _, v := range e.EligibleVoters {
		err := vw.SQLDB.InsertGenericTX(&v, tx)
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

func (vw *VoteWatcher) AddEligibleVoter(list *EligibleList, voter *EligibleVoterEntry, hash [32]byte) error {
	if vw.UseMemory {
		list.SubmittedEntries[hash] = true

		err := list.AddVoter(voter)
		return err
	}

	tx, err := vw.SQLDB.Begin()
	if err != nil {
		return err
	}

	fmt.Println("INSERT A VOTER!!!!!!!!!!!")
	for _, v := range voter.Content {
		fmt.Println("INSERT VOTER", v.VoterID.String())
		err := vw.SQLDB.InsertGenericTX(&v, tx)
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
