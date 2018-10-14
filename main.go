package main

import (
	"flag"
	_ "flag"
	"fmt"

	"github.com/FactomProject/factomd/common/primitives"

	"io/ioutil"
	"strings"

	"github.com/Emyrk/go-factom-vote/vote"
	log "github.com/sirupsen/logrus"
)

func main() {
	var (
		all     = flag.Bool("all", false, "Parse for all identities")
		rootHex = flag.String("v", "", "Vote Chain in hex")
		factomd = flag.String("s", "localhost:8088", "Factomd api location")
		pretty  = flag.Bool("p", false, "Make the printout pretty for us mere humans")
		loglvl  = flag.String("l", "none", "Set log level to 'debug', 'info', 'warn', 'error', or 'none'")
	)

	flag.Parse()

	switch strings.ToLower(*loglvl) {
	case "warn", "warning":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "none":
		log.SetLevel(log.FatalLevel)
		log.SetOutput(ioutil.Discard)
	}

	var data []byte
	c := vote.NewAPIController(*factomd)
	if !c.IsWorking() {
		fmt.Println("Factomd location is not working")
		return
	}

	if *all {
		//ids, err := parseAll(c)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//
		//data, err = json.Marshal(ids)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
	} else {
		// Single
		if *rootHex == "" {
			fmt.Println("go-factom-vote -id=888888....")
			return
		}

		v, err := parseSingle(*rootHex, c)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(v)
	}

	var _ = pretty
	var _ = data
	//if *pretty {
	//	var dst bytes.Buffer
	//	err := json.Indent(&dst, data, "", "\t")
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	fmt.Println(string(dst.Bytes()))
	//} else {
	//	fmt.Println(string(data))
	//}
}

func parseSingle(votechainHex string, c *vote.Controller) (*vote.Vote, error) {
	votechain, err := primitives.HexToHash(votechainHex)
	if err != nil {
		return nil, fmt.Errorf("parsing vote chain id: %s", err.Error())
	}

	vote, err := c.FindVote(votechain)
	return vote, err
}

//func parseAll(c *factom_identity.Controller) (map[string]*identity.Identity, error) {
//	ids, err := c.FindAllIdentities()
//	return ids, err
//}
