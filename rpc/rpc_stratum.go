package rpc

import (
	//"bytes"
	//"crypto/sha256"
	"encoding/json"
	"errors"
	//"fmt"
	"math/big"
	//"net/http"
	"net"
	//"net/url"
	//"strings"
	"sync"
	"log"
	"time"
	"strconv"
	"bufio"
	"io"
	//"fmt"
	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Tomahna81/ethash-mining-stratum-proxy/util"
	//"github.com/Tomahna81/ethash-mining-stratum-proxy/proxy"
)

type StratumReq struct {
	Id     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	//Params []string `json:"params"`
	Worker string `json:"worker"`
}

type StratumRes struct {
	Id      int64       `json:"id"`
	Version string      `json:"jsonrpc"`
	//Result  interface{} `json:"result"`
	Result  []string `json:"result"`
}

// Used for the first login
type StratumRes_2 struct {
    Id      int64       `json:"id"`
    Version string      `json:"jsonrpc"`
    //Result  interface{} `json:"result"`
    Result  bool `json:"result"`
}

type RPCStratumClient struct {
	sync.RWMutex
	Url         string
	Host        string
	Port        int
	Name        string
	sick        bool
	sickRate    int
	successRate int
	client      net.Conn
    enc  *json.Encoder
    dec  *json.Decoder
}


/*type GetBlockReply struct {
	Number       string   `json:"number"`
	Hash         string   `json:"hash"`
	Nonce        string   `json:"nonce"`
	Miner        string   `json:"miner"`
	Difficulty   string   `json:"difficulty"`
	GasLimit     string   `json:"gasLimit"`
	GasUsed      string   `json:"gasUsed"`
	Transactions []Tx     `json:"transactions"`
	Uncles       []string `json:"uncles"`
	// https://github.com/ethereum/EIPs/issues/95
	SealFields []string `json:"sealFields"`
}
*/

/*
type GetBlockReplyPart struct {
	Number     string `json:"number"`
	Difficulty string `json:"difficulty"`
}
*/

//const receiptStatusSuccessful = "0x1"

/*
type TxReceipt struct {
	TxHash    string `json:"transactionHash"`
	GasUsed   string `json:"gasUsed"`
	BlockHash string `json:"blockHash"`
	Status    string `json:"status"`
}
*/

/*
func (r *TxReceipt) Confirmed() bool {
	return len(r.BlockHash) > 0
}


// Use with previous method
func (r *TxReceipt) Successful() bool {
	if len(r.Status) > 0 {
		return r.Status == receiptStatusSuccessful
	}
	return true
}
*/

/*
type Tx struct {
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Hash     string `json:"hash"`
}
*/
/*
type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}
*/

func initStratumClient(name string, Address string, StratumPort int) *RPCStratumClient {
    ip := net.ParseIP(Address)
    raddr := &net.TCPAddr{
        IP:   net.IP(ip),
        Port: StratumPort,
    }
    conn, err := net.DialTCP("tcp4", nil, raddr)
    if err != nil {
        println(" Error: ", err.Error())
    }

    enc := json.NewEncoder(conn)
    dec := json.NewDecoder(conn)
    nc := &RPCStratumClient{
    	Name: name,
        Host: Address,
        Port: StratumPort,
        client:    conn,
        enc:  enc,
        dec:  dec,
    }
    return nc
}

func NewRPCStratumClient(name string, host string , port string, username string, timeout string) *RPCStratumClient { //(*RPCStratumClient, error) {
	//u, err1 := url.Parse(fullurl)
	//host, port, err2 := net.SplitHostPort(u.Host)
	//log.Println("u.Host = ", u.Host )
	//if err1 != nil || err2 != nil{
		//err:=errors.New("The provided URL for the upstram is not in the correct format. Please check that you provide the address and port adequatly")
	//	log.Printf("The provided URL for the upstram is not in the correct format. Please check that you provide the address and port adequatly")
		//var dummy *RPCStratumClient
		//return dummy//, err
	//}
	//rpcClient := &RPCClient{Name: name, Url: url}
	port_int, _:=strconv.Atoi(port)
	rpcStratum:=initStratumClient(name, host, port_int)
	rpcStratum.Url=host
	timeoutIntv := util.MustParseDuration(timeout)
	rpcStratum.client.SetDeadline(time.Now().Add(timeoutIntv))
	params := []string{username, ""}
	reply, err_login:=rpcStratum.StratumSubmitLogin(params)
	log.Println(" Reply at first login: ", reply)
	if err_login != nil{
		log.Println("    ERROR: Could not login into the pool", err_login)
	}
	//log.Printf(" Error at first login: ", err_login)
	return rpcStratum//,err_login
}


func (r *RPCStratumClient) StratumSubmitLogin(params []string) (string, error){
        //var params []string
        //var req StratumReq
        //err := json.Unmarshal(req.Params, &params)
        //if err != nil {
           // err=errors.New("Malformed stratum request params to be sent upstream: Check the params formating")
        //    return "", err
        //} else{
            jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": "eth_submitLogin", "params": params, "id": 0}//, "worker": worker}
            data, _ := json.Marshal(jsonReq)
            data_str:=string(data) + "\n"
            //log.Println(".   raw data sent:", data_str)
            reply, err:=r.StratumRequest(data_str) // Submit request and process response. The response is a string here
           	if err != nil {
				return "", err
			}
            return reply, nil
        //}
}

func (r *RPCStratumClient) StratumGetWork() ([]string, error) {
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": "eth_getWork", "id": 0} // The getwork request
	data, _ := json.Marshal(jsonReq) // Ensure proper formating
    data_str:=string(data) + "\n" // Add terminator
	reply, err:=r.StratumRequest(data_str) // Submit request and process response. The response is a string here
	if err != nil {
		return []string{""}, err
	}
    var ans StratumRes
   if err == nil{ 
        //println(" We sent: ", req)
        err:=json.Unmarshal([]byte(reply), &ans)
        if err !=nil{
        	log.Println("Error in formating Json string inside StratumGetWork : ", err.Error())
        	log.Println(" Raw Message:", reply)
        }
        /*if err == nil {
            log.Println(" We got a reply: ", reply)
            log.Println("  Formated output: ")
            log.Println("    ID: ", ans.Id) 
            log.Println("    jsonrpc: ", ans.Version)
            log.Println(".   result: ", ans.Result[0])
            log.Println(".   result: ", ans.Result[1])
            log.Println(".   result: ", ans.Result[2])
            if len(ans.Result) == 4{
            	log.Println(".   result: ", ans.Result[3])
            }
        }
        */
    } else{
        log.Println("Error: Could not format the retrieved work from the pool")
    }
	return ans.Result, err
}

func (r *RPCStratumClient) StratumGetPendingBlock() (*GetBlockReplyPart, error) {
	errors.New("WARNING:  GetPendingBlock is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumGetBlockByHeight(height int64) (*GetBlockReply, error) {
	//params := []interface{}{fmt.Sprintf("0x%x", height), true}
	errors.New("WARNING:  GetBlockByHeight is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumGetBlockByHash(hash string) (*GetBlockReply, error) {
	errors.New("WARNING:  GetBlockByHash is not supported yet in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumGetUncleByBlockNumberAndIndex(height int64, index int) (*GetBlockReply, error) {
	errors.New("WARNING: GetUncleByBlockNumberAndIndex is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumgetBlockBy(method string, params []interface{}) (*GetBlockReply, error) {
	errors.New("WARNING: GetBlockBy is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumGetTxReceipt(hash string) (*TxReceipt, error) {
	errors.New("WARNING: GetTxReceipt is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumSubmitBlock(params []string) (uint8, []string, error) {
	jsonReq := map[string]interface{}{"jsonrpc": "2.0", "method": "eth_submitWork", "params": params, "id": 0}//, "worker": worker}
	data, _ := json.Marshal(jsonReq) // Ensure proper formating
    data_str:=string(data) + "\n" // Add terminator
	reply, err:=r.StratumRequest(data_str) // Submit request and process response. The response is a string here
	if err != nil {
		return 0, []string{}, err
	}
//	// We test first the case where the pool answers with a boolean for the result
//    var ans2 StratumRes_2
//    err=json.Unmarshal([]byte(data_str), &ans2)
//    if err != nil{
//   		// If fails, consider that the answer is a new job
       var ans StratumRes
       err=json.Unmarshal([]byte(reply), &ans)
       if err != nil{
 		 log.Println("Error: Could not parse the pool answer after share submission")
 		 log.Println("Raw Message from the pool: ", data_str)
 		 return 0, []string{""}, err 
       }
//    } else{
//    	b ,_:= strconv.ParseBool(fmt.Sprintf("%t", ans2.Result))
//    	if b == true{
//    		return 1, []string{""}, nil
//    	} else{
//    		return 0, []string{""}, nil
//    	}	
//    }

/*	var ans StratumRes
    err=json.Unmarshal([]byte(reply), &ans)
    if err != nil{
    	b ,_:= strconv.ParseBool(fmt.Sprintf("%t", ans.Result))
    	//log.Println(". Parsed boolean : ", b)
    	return b, err
	}
*/
	//return 0, []string{""}, err
    log.Println(" ------ SUBMITTING SHARE INFO ------")
    log.Println(" Submission: ", data_str)
    log.Println(" Reply     : ", reply)
    log.Println(" Errors    : ", err)
    log.Println(" output.   : ", ans.Result)
    log.Println(" output.   : ", ans.Result[0])
    log.Println(" output.   : ", ans.Result[1])
    return 2, ans.Result, nil // 2 is for an answer that is a job
}

func (r *RPCStratumClient) StratumGetBalance(address string) (*big.Int, error) {
	errors.New("WARNING: eth_getbalance is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return nil, nil
}

func (r *RPCStratumClient) StratumSign(from string, s string) (string, error) {
	errors.New("WARNING: eth_sign is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return "", nil
}

func (r *RPCStratumClient) StratumGetPeerCount() (int64, error) {
	errors.New("WARNING: eth_sign is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return -1, nil
}

func (r *RPCStratumClient) StratumGetGasPrice() (int64, error) {
	errors.New("WARNING: eth_getgasprice is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return -1, nil
}

func (r *RPCStratumClient) SendTransaction(from, to, gas, gasPrice, value string, autoGas bool) (string, error) {
	errors.New("WARNING: sending transactions is not supported in ETH-PROXY Stratum mode. This will be ignored")
	return "", nil
}

func (r *RPCStratumClient) StratumRequest(data string) (string, error) {
    _, err := r.client.Write([]byte(data))
     if err != nil {
        //log.Println("Error when requesting some work from the upstream (eth_getWork)")
        r.markSick()
        err=errors.New("Error when requesting some work from the upstream (eth_getWork)")
        return "", err
    } else{
        reply, errReply :=r.HandleRead()
        if errReply != nil {
            err=errors.New("Error while reading the reply from the upstream")
            return "", errReply
        }
        return reply, nil
    }
}

func (r *RPCStratumClient) HandleRead() (response string, err error) {
	MaxReqSize := 1024
    time_string:="10s"
    timeout, _:= time.ParseDuration(time_string)
    r.client.SetDeadline(time.Now().Add(timeout))
    for {
        connbuff := bufio.NewReaderSize(r.client, MaxReqSize)
        data, isPrefix, err := connbuff.ReadLine()
        if isPrefix == true {
        	r.markSick()
        	log.Printf("isPrefix is true... Issue: Truncation of the server response occured!")
            err=errors.New("isPrefix is true... Issue: Truncation of the server response occured!")
            return "", err
        } 
        if err != nil {
            if err == io.EOF {
            	r.markSick()
                log.Printf("We got disconnected from the upstream")
                //errors.New("We got disconnected from the upstream")
                break
            } 
            if err != nil {
            	r.markSick()
            	log.Printf("Error reading from socket: %v", err.Error())
                //err=errors.New("Error reading from socket")
                return "", err
            }
        }
        if len(data) > 1 {
            response, err:= handleresponse(data)
            //var expected_response StratumRes
            //err = json.Unmarshal(data, &expected_response)
            if err != nil {
                println("Malformed stratum request params to be sent upstream: Check the formating of the response: ", err.Error())
                println(".  Raw response:     ", string(data))
                return "", err
            }
            response=string(data)
            return response, nil
            break
            //cl.conn.SetDeadline(time.Now().Add(timeout))
            //s.setDeadline(cl.conn) // For the case of implemation inside the proxy
            
            //if err != nil {
            //    return "", err
            //}
        }
    }
    return response, nil
}

func (r *RPCStratumClient) Check() bool {
	_, err := r.StratumGetWork()
	if err != nil {
		return false
	}
	r.markAlive()
	return !r.Sick()
}

func (r *RPCStratumClient) Sick() bool {
	r.RLock()
	defer r.RUnlock()
	return r.sick
}

func (r *RPCStratumClient) markSick() {
	r.Lock()
	r.sickRate++
	r.successRate = 0
	if r.sickRate >= 5 {
		r.sick = true
	}
	r.Unlock()
}

func (r *RPCStratumClient) markAlive() {
	r.Lock()
	r.successRate++
	if r.successRate >= 5 {
		r.sick = false
		r.sickRate = 0
		r.successRate = 0
	}
	r.Unlock()
}

func handleresponse(data []byte) (response string, err error){
    var expected_response StratumRes
    err = json.Unmarshal(data, &expected_response)
    if err != nil {
        var expected_response_2 StratumRes_2
        err = json.Unmarshal(data, &expected_response_2)
        if err != nil {
            println("Stratum RPC Error: Malformed stratum request params to be sent upstream: Check the formating of the response ", err.Error())
            println("     Raw response: ", string(data))
            return "", err
        }
    } 
    response=string(data)
    return response, nil
}

