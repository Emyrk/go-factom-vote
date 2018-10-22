package vote

// All vote modifications go through here
func (vw *VoteWatcher) AddNewVoteProposal(p *Vote) {
	vw.Lock()
	vw.VoteProposals[p.Proposal.ProposalChain.Fixed()] = p
	vw.Unlock()
}

func (vw *VoteWatcher) AddReveal(p *Vote, r VoteReveal, height uint32) error {
	return p.AddReveal(r, height)
}

func (vw *VoteWatcher) AddCommit(p *Vote, c VoteCommit, height uint32) error {
	return p.AddCommit(c, height)
}

func (vw *VoteWatcher) SetRegistered(p *Vote, registered bool) error {
	p.Registered = registered
	return nil
}
