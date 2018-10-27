package vote

import (
	"fmt"

	. "github.com/Emyrk/go-factom-vote/vote/common"
)

// All vote modifications go through here
func (vw *VoteWatcher) AddNewVoteProposal(v *Vote) {
	vw.Lock()
	vw.VoteProposals[v.Proposal.ProposalChain.Fixed()] = v
	err := vw.SQLDB.InsertVote(v)
	if err != nil {
		fmt.Println(err)
	}
	vw.Unlock()
}

func (vw *VoteWatcher) AddReveal(v *Vote, r VoteReveal, height uint32) error {
	return v.AddReveal(r, height)
}

func (vw *VoteWatcher) AddCommit(v *Vote, c VoteCommit, height uint32) error {
	return v.AddCommit(c, height)
}

func (vw *VoteWatcher) SetRegistered(v *Vote, registered bool) error {
	v.Registered = registered
	return nil
}
