package scraper

import (
	"fmt"

	"time"

	"github.com/Emyrk/go-factom-vote/vote"
	"github.com/Emyrk/go-factom-vote/vote/common"
	"github.com/Emyrk/go-factom-vote/vote/database"
	"github.com/FactomProject/factomd/common/interfaces"
	log "github.com/sirupsen/logrus"
)

var scraperlog = log.WithFields(log.Fields{"file": "scraper.go"})

type Scraper struct {
	Factom          Fetcher
	Database        *database.SQLDatabase
	WalletdLocation string

	// IdentityControl
	VoteControl *vote.VoteWatcher
}

func NewScraper(host string, port int, config *database.SqlConfig) (*Scraper, error) {
	flog := scraperlog.WithField("func", "NewScraper")

	s := new(Scraper)
	factomd := fmt.Sprintf("%s:%d", host, port)
	s.Factom = NewAPIReader(factomd)
	_, err := s.Factom.FetchDBlockHead()
	if err != nil {
		return nil, err
	}
	flog.Infof("Factomd location %s", factomd)

	if config != nil {
		s.Database, err = database.InitDb(*config)
	} else {
		s.Database, err = database.InitLocalDB()
	}
	if err != nil {
		return nil, err
	}
	flog.Infof("Postgres database connected")

	s.WalletdLocation = "localhost:8089"

	s.VoteControl = vote.NewVoteWatcherWithDB(s.Database)
	// TODO: Sync Vote Control

	return s, nil
}

var CurrentCatchup uint32
var CurrentTop uint32

func (s *Scraper) Catchup() {
	flog := scraperlog.WithFields(log.Fields{"func": "CatchUp"})
	flog.Info("Catchup started")
	// Find the highest height completed
	next := uint32(s.Database.FetchHighestDBInserted() + 1)

	getNextTop := func() uint32 {
		for {
			topDblock, err := s.Factom.FetchDBlockHead()
			if err != nil {
				flog.Error(err)
				time.Sleep(3 * time.Second)
				continue
			}
			return topDblock.GetDatabaseHeight()
		}
	}

	start := time.Now()
	top := getNextTop()
	CurrentTop = top
	changes := 0

MainCatchupLoop:
	for {
		if next%10 == 0 {
			flog.WithFields(log.Fields{"current": next, "top": top, "time": time.Since(start), "changes": changes}).Info("")
		}
		start = time.Now()
		if next > top {
			top = getNextTop()
			if next > top {
				time.Sleep(30 * time.Second)
				continue
			}
		}
		CurrentCatchup = next

		dblock, err := s.Factom.FetchDBlockByHeight(next)
		if err != nil {
			errorAndWait(flog.WithField("fetch", "dblock"), err)
			continue
		}

		var eblocks []interfaces.IEntryBlock
		for _, e := range dblock.GetEBlockDBEntries() {
			eblock, err := s.Factom.FetchEBlock(e.GetKeyMR())
			if err != nil {
				errorAndWait(flog.WithField("fetch", "eblock"), err)
				continue MainCatchupLoop
			}
			eblocks = append(eblocks, eblock)
		}

		// Parse Entries

		height := dblock.GetDatabaseHeight()
		hog := flog.WithField("height", height)
		for _, e := range eblocks {
			t := dblock.GetHeader().GetTimestamp().GetTime()
			//nt := t
			// Parse all entries
			for _, ehash := range e.GetEntryHashes() {
				if ehash.IsMinuteMarker() {
					//nt = t.Add(time.Minute * time.Duration(common.MinuteMarkerToInt(ehash)))
					continue
				}

				entry, err := s.Factom.FetchEntry(ehash.String())
				if err != nil {
					errorAndWait(hog.WithFields(log.Fields{"fetch": "entry", "hash": ehash.String()}), err)
					continue MainCatchupLoop
				}
				change, err := s.VoteControl.ProcessEntry(entry, height, t, true)
				if err != nil {
					hog.WithFields(log.Fields{"vote-parse": "entry", "hash": ehash.String()}).Error(err)
					//continue MainCatchupLoop // TODO :Remove
				}
				if change {
					changes++
				}
			}
		}

		s.VoteControl.ProcessOldEntries()

		// Now we check if any votes are complete
		err = s.computeResults(int(height))
		if err != nil {
			errorAndWait(hog.WithFields(log.Fields{"insert": "completed"}), err)
			continue MainCatchupLoop
		}

		err = s.Database.InsertCompleted(int(next))
		if err != nil {
			errorAndWait(hog.WithFields(log.Fields{"insert": "completed"}), err)
			continue MainCatchupLoop
		}
		// End loop
		next++
		changes = 0
	}
}

func (s *Scraper) computeResults(dbheight int) error {
	flog := scraperlog.WithFields(log.Fields{"func": "computeResults", "height": dbheight})
	votes, err := s.Database.FetchCompleteVotes(dbheight)
	if err != nil {
		return err
	}

	tx, err := s.Database.Begin()
	if err != nil {
		return err
	}

	for _, v := range votes {
		// Grab voters, commits, and reveals
		voters, err := s.Database.FetchEligibleVoters(v.Proposal.Vote.EligibleVotersChainID.String(), v.Proposal.Vote.PhasesBlockHeights.CommitStart)
		if err != nil {
			tx.Rollback()
			return err
		}

		commits, err := s.Database.FetchCommits(v.Proposal.ProposalChain.String())
		if err != nil {
			tx.Rollback()
			return err
		}

		var _ = commits

		reveals, err := s.Database.FetchReveals(v.Proposal.ProposalChain.String())
		if err != nil {
			tx.Rollback()
			return err
		}

		results, err := common.ComputeResult(v, voters, reveals)
		if err != nil {
			flog.WithField("vote", v.Proposal.ProposalChain.String()).Error(err)
		}

		if results != nil {
			err = s.Database.InsertGenericTX(results, tx)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func errorAndWait(logger *log.Entry, err error) {
	logger.Error(err)
	time.Sleep(2 * time.Second)
}
