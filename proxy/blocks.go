package proxy

import (
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"os"
	"github.com/ethereum/go-ethereum/common"
	"github.com/Tomahna81/ethash-proxy/rpc"
	"github.com/Tomahna81/ethash-proxy/util"
)

const maxBacklog = 3

type heightDiffPair struct {
	diff   *big.Int
	height uint64
}

type BlockTemplate struct {
	sync.RWMutex
	Header               string
	Seed                 string
	Target               string
	Difficulty           *big.Int
	Height               uint64
	GetPendingBlockCache *rpc.GetBlockReplyPart
	nonces               map[string]bool
	headers              map[string]heightDiffPair
}

type Block struct {
	difficulty  *big.Int
	hashNoNonce common.Hash
	nonce       uint64
	mixDigest   common.Hash
	number      uint64
}

func (b Block) Difficulty() *big.Int     { return b.difficulty }
func (b Block) HashNoNonce() common.Hash { return b.hashNoNonce }
func (b Block) Nonce() uint64            { return b.nonce }
func (b Block) MixDigest() common.Hash   { return b.mixDigest }
func (b Block) NumberU64() uint64        { return b.number }

func (s *ProxyServer) fetchBlockTemplate() {
	rpc_strat := s.rpc_stratum() // Mod Tomahna
	//pendingReply, height, diff, err := s.fetchPendingBlock() // Mod Tomahna
	//if err != nil {
	//	log.Printf("Error while refreshing pending block on %s: %s", rpc_stratum.Name, err)
	//	return
	//} // End Mod Tomahna
	reply, err := rpc_strat.StratumGetWork()
	if err != nil {
		log.Printf("Error while refreshing block template on %s: %s", rpc_strat.Name, err)
		s.markSick()
		log.Printf("     ==> BUG TO BE FIXED: FORCING EXIT")
		os.Exit(3)
		return
	} else{
		s.markOk()
		t := s.currentBlockTemplate()
		if len(reply) <3{
			return
		}
		// No need to update, we have fresh job
		if t != nil && t.Header == reply[0] { // Mod Tomahna
			return
		}
		// If we see 4 outputs, we consider the 4th one as providing the height
		/* check if the block number is in a valid range A year has ~31536000 seconds 50 years have ~1576800000
           assuming a (very fast) blocktime of 10s: ==> in 50 years we get 157680000 (=0x9660180) blocks
        */
        var height uint64
        height=0
        if len(reply) ==4{
        	height, _ = strconv.ParseUint(strings.Replace(reply[3], "0x", "", -1), 16, 64)
    	}
        if height > 157680000 {
        	//log.Println(" ===> Height modfified to 0")
			height= 0      
        }
 
		//pendingReply.Difficulty = util.ToHex(s.config.Proxy.Difficulty)
		// Mod Tomahna to ensure compatibility (this might not be used in the proxy)
		var pendingReply *rpc.GetBlockReplyPart
		//pendingReply :=*rpc.GetBlockReplyPart{ // This crashes I do not know why
		//	Number :    "0",
		//	Difficulty : reply[2],
		//}
	
		newTemplate := BlockTemplate{
			Header:               reply[0],
			Seed:                 reply[1],
			Target:               reply[2],
			Height:               height,
			//Difficulty:           big.NewInt(diff), // Mod Tomahna
			Difficulty:           util.TargetHexToDiff(reply[2]), // Mod Tomahna
			GetPendingBlockCache: pendingReply,
			headers:              make(map[string]heightDiffPair),
		}
		// Copy job backlog and add current one
		newTemplate.headers[reply[0]] = heightDiffPair{
			diff:   util.TargetHexToDiff(reply[2]),
			height: height,
		}
		if t != nil {
			for k, v := range t.headers {
				if v.height > height-maxBacklog {
					newTemplate.headers[k] = v
				}
			}
		}
		s.blockTemplate.Store(&newTemplate)
		log.Println(" height: ", height)
		log.Printf("New block to mine on %s at height %d / %s . Difficulty: %d", rpc_strat.Name, height, reply[0][0:10], util.TargetHexToDiff(reply[2]))
		//log.Println(" Difficulty set to: ", util.TargetHexToDiff(reply[2]))
    	diff_new:=util.TargetHexToDiff(reply[2])
    	s.diff=diff_new.String()	// Stratum
		if s.config.Proxy.Stratum.Enabled {
			go s.broadcastNewJobs()
		}
	
		if s.config.Proxy.StratumSSL.Enabled {
			go s.broadcastSSLNewJobs()
		}
	
		if s.config.Proxy.StratumNiceHash.Enabled {
			go s.broadcastNewJobsNH()
		}
	}
}


func (s *ProxyServer) fetchBlockTemplateV2(reply []string) {
	rpc_strat := s.rpc_stratum() // Mod Tomahna
	//pendingReply, height, diff, err := s.fetchPendingBlock() // Mod Tomahna
	//if err != nil {
	//	log.Printf("Error while refreshing pending block on %s: %s", rpc_stratum.Name, err)
	//	return
	//} // End Mod Tomahna
	//reply, err := rpc_strat.StratumGetWork()
	//if err != nil {
	//	log.Printf("Error while refreshing block template on %s: %s", rpc_strat.Name, err)
	//	return
	//}
	
	t := s.currentBlockTemplate()
		// No need to update, we have fresh job
	if t != nil && t.Header == reply[0] { // Mod Tomahna
		return
	}
	// If we see 4 outputs, we consider the 4th one as providing the height
		/* check if the block number is in a valid range A year has ~31536000 seconds 50 years have ~1576800000
           assuming a (very fast) blocktime of 10s: ==> in 50 years we get 157680000 (=0x9660180) blocks
        */
        var height uint64
        height=0
        if len(reply) ==4{
        	height, _ = strconv.ParseUint(strings.Replace(reply[3], "0x", "", -1), 16, 64)
    	}
        if height > 157680000 {
        	//log.Println(" ===> Height modfified to 0")
			height= 0      
        }
 
	//pendingReply.Difficulty = util.ToHex(s.config.Proxy.Difficulty)
	// Mod Tomahna to ensure compatibility (this might not be used in the proxy)
	var pendingReply *rpc.GetBlockReplyPart
	//pendingReply :=*rpc.GetBlockReplyPart{ // This crashes I do not know why
	//	Number :    "0",
	//	Difficulty : reply[2],
	//}

	newTemplate := BlockTemplate{
		Header:               reply[0],
		Seed:                 reply[1],
		Target:               reply[2],
		Height:               height,
		//Difficulty:           big.NewInt(diff), // Mod Tomahna
		Difficulty:           util.TargetHexToDiff(reply[2]), // Mod Tomahna
		GetPendingBlockCache: pendingReply,
		headers:              make(map[string]heightDiffPair),
	}
	// Copy job backlog and add current one
	newTemplate.headers[reply[0]] = heightDiffPair{
		diff:   util.TargetHexToDiff(reply[2]),
		height: height,
	}
	if t != nil {
		for k, v := range t.headers {
			if v.height > height-maxBacklog {
				newTemplate.headers[k] = v
			}
		}
	}
	s.blockTemplate.Store(&newTemplate)
	log.Println(" height: ", height)
	log.Printf("New block to mine on %s at height %d / %s . Difficulty: %d", rpc_strat.Name, height, reply[0][0:10], util.TargetHexToDiff(reply[2]))
	//log.Println(" Difficulty set to: ", util.TargetHexToDiff(reply[2]))
    diff_new:=util.TargetHexToDiff(reply[2])
    s.diff=diff_new.String()	// Stratum
	if s.config.Proxy.Stratum.Enabled {
		go s.broadcastNewJobs()
	}

	if s.config.Proxy.StratumSSL.Enabled {
		go s.broadcastSSLNewJobs()
	}

	if s.config.Proxy.StratumNiceHash.Enabled {
		go s.broadcastNewJobsNH()
	}
}

