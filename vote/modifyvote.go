package vote

import . "github.com/Emyrk/go-factom-vote/vote/common"

// All vote modifications go through here
func (vw *VoteWatcher) AddNewVoteProposal(v *Vote) error {
	//vw.Lock()
	vw.VoteProposals[v.Proposal.ProposalChain.Fixed()] = v
	err := vw.SQLDB.InsertVote(v)
	return err
	//vw.Unlock()
}

func (vw *VoteWatcher) AddReveal(v *Vote, r VoteReveal, height uint32) error {
	err := vw.SQLDB.InsertGeneric(&r)
	if err != nil {
		return err
	}
	return v.AddReveal(r, height)
}

func (vw *VoteWatcher) AddCommit(v *Vote, c VoteCommit, height uint32) error {
	err := vw.SQLDB.InsertGeneric(&c)
	if err != nil {
		return err
	}
	return v.AddCommit(c, height)
}

func (vw *VoteWatcher) SetRegistered(v *Vote, registered bool) error {
	v.Registered = registered
	return vw.SQLDB.SetRegistered(v.Proposal.ProposalChain)
}

func (vw *VoteWatcher) AddNewEligibleList(e *EligibleList) error {
	err := vw.SQLDB.InsertGeneric(e)
	if err != nil {
		return err
	}

	vw.EligibleLists[e.ChainID.Fixed()] = e
	return nil
}

func (vw *VoteWatcher) AddEligibleVoter(list *EligibleList, voter *EligibleVoterEntry, hash [32]byte) error {
	tx, err := vw.SQLDB.Begin()
	if err != nil {
		return err
	}

	for _, v := range voter.Content {
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

	err = list.AddVoter(voter)
	if err != nil {
		return err
	}

	list.SubmittedEntries[hash] = true
	return nil
}
