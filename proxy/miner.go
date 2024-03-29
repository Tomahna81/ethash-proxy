package proxy

import (
	"log"
	//"math/big"
	//"strconv"
	//"strings"

	"github.com/ethereum/ethash"
	//"github.com/ethereum/go-ethereum/consensus/ethash"
	//"github.com/ethereum/go-ethereum/common"
)

var hasher = ethash.New()
//var hasher = ethash.init()

func (s *ProxyServer) processShare(login, id, ip string, t *BlockTemplate, params []string) (bool, bool) {
	//nonceHex := params[0]
	hashNoNonce := params[1]
	//mixDigest := params[2]
	//nonce, _ := strconv.ParseUint(strings.Replace(nonceHex, "0x", "", -1), 16, 64)
	shareDiff := s.config.Proxy.Difficulty

	h, ok := t.headers[hashNoNonce]
	if !ok {
		//TODO:Store stale share in Redis
		log.Printf("Stale share from %v@%v", login, ip)
		return false, false
	}

	//share := Block{
	//	number:      h.height,
	//	hashNoNonce: common.HexToHash(hashNoNonce),
	//	difficulty:  big.NewInt(shareDiff),
	//	nonce:       nonce,
	//	mixDigest:   common.HexToHash(mixDigest),
	//}
	//if !hasher.Verify(share) {
	//	return false, false
	//}

	//Write the Ip address into the settings:login:ipaddr and timeit added to settings:login:iptime hash
	s.backend.LogIP(login,ip)

	//if hasher.Verify(block) {
	okshare, result, err := s.rpc_stratum().StratumSubmitBlock(params)
	log.Println(" Submitted Share: ", params)
	//log.Println(" Status :", okshare)
	//log.Println(" Errors :", err)
	if err != nil {
		log.Printf("Share submission failure at height %v for %v: %v", h.height, t.Header, err)
	} else if okshare == 0 {
		log.Printf("Share rejected at height %v for %v", h.height, t.Header)
		return false, false
	} else {
		if okshare == 1{ // We just got a boolean from the pool and need to request a job before fetching it
			s.fetchBlockTemplate()
		}
		if okshare == 2{ // We got a reply from the pool that provides a new job
			s.fetchBlockTemplateV2(result)				
		}
		exist, err := s.backend.WriteShare(login, id, params, shareDiff, h.height, s.hashrateExpiration)
		if exist {
			return true, false
		}
		if err != nil {
			log.Println("Failed to insert share data into backend:", err)
		}
	}
	return false, true
}
