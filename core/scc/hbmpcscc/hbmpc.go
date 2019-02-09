/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hbmpcscc

import (
	// XXX for HBMPC
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger/fabric/core/aclmgmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/ledger"
	"github.com/hyperledger/fabric/core/peer"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
)

func backgroundTask(host string, port string) {
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
	}
	defer l.Close()
	fmt.Println("Listening on " + host + ":" + port)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	fmt.Println("In handle request")
	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	} else {
		fmt.Println("Received : ", buf, reqLen)
	}

	conn.Write([]byte("Message received."))
	conn.Close()
}

func send(server string, port string) {

	// connect to this socket
	conn, err := net.Dial("tcp", server+":"+port)
	if err != nil {
		fmt.Println("error connecting:", err.Error())
	} else {
		fmt.Print("Text to send: ")
		text := "hello"
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Message from server: " + message)
	}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// New returns an instance of HBMPCSCC.
// Typically this is called once per peer.
func New(aclProvider aclmgmt.ACLProvider) *MpcEngine {
	// XXX ---- HoneyBadgerMPC ----
	//
	// From https://github.com/initc3/HoneyBadgerMPC/issues/24#issuecomment-417183754:
	//
	// Identify a suitable "snip" point in Hyperledger Fabric endorser code, that lets
	// endorsers communicate with the hbmpc application before approving a transaction.
	// This should be at the endorser, roughly in between when it receives a transaction
	// and when it signs it. Essentially the mpc will act like a "pre-ordering" service
	//
	// TODO Replace "dummy" code below with appropriate code.
	//cmd := exec.Command("python3", "-m", "honeybadgermpc.polynomial")
	fmt.Println("--------------------------------------------------------------------\n")
	fmt.Println("net.LookupHost peer0.org1.example.com")
	peer0org1, err := net.LookupHost("peer0.org1.example.com")
	if err != nil {
		log.Fatalf("net.LookupHost peer0.org1.example.com failed with %s\n", err)
	}
	fmt.Println(peer0org1)
	fmt.Println(peer0org1[0])
	fmt.Println("New change testing 1 ....")
	fmt.Println("\n--------------------------------------------------------------------")
	fmt.Println("--------------------------------------------------------------------\n")
	fmt.Println("net.LookupHost peer1.org1.example.com")
	peer1org1, err := net.LookupHost("peer1.org1.example.com")
	if err != nil {
		log.Fatalf("net.LookupHost peer1.org1.example.com failed with %s\n", err)
	}
	fmt.Println(peer1org1)
	fmt.Println(peer1org1[0])
	fmt.Println("\n--------------------------------------------------------------------")
	go backgroundTask(GetOutboundIP().String(), "9001")
	ncservecmd := exec.Command("nc", "-l", "-p", "9001")
	errmsg := ncservecmd.Start()
	if errmsg != nil {
		fmt.Printf("ncservecmd.Start() failed with %s\n", errmsg)
	}
	//fmt.Printf("HoneyBadgerMPC ... :\n%s\n", string(out))
	//fmt.Printf("netcat cmd output, `nc -l -p 9000`:\n%s\n", string(out))

	time.Sleep(1000 * time.Millisecond)
	fmt.Println("Sending Message ")
	send(peer0org1[0], "9001")
	time.Sleep(5000 * time.Millisecond)
	echocmd := exec.Command("echo", "hi", "|", "nc", peer0org1[0], "9001")
	echocmderrmsg := echocmd.Start()
	if echocmderrmsg != nil {
		fmt.Printf("echocmd.Run() failed with %s\n", echocmderrmsg)
	}
	/*
		for {
		}
	*/
	// -XXX --- HoneyBadgerMPC ----

	return &MpcEngine{
		aclProvider: aclProvider,
	}
}

func (e *MpcEngine) Name() string              { return "hbmpcscc" }
func (e *MpcEngine) Path() string              { return "github.com/hyperledger/fabric/core/scc/hbmpcscc" }
func (e *MpcEngine) InitArgs() [][]byte        { return nil }
func (e *MpcEngine) Chaincode() shim.Chaincode { return e }
func (e *MpcEngine) InvokableExternal() bool   { return true }
func (e *MpcEngine) InvokableCC2CC() bool      { return true }
func (e *MpcEngine) Enabled() bool             { return true }

// MpcEngine implements the MPC functions, including:
// - GetChainInfo returns BlockchainInfo
// - GetBlockByNumber returns a block
// - GetBlockByHash returns a block
// - GetTransactionByID returns a transaction
type MpcEngine struct {
	aclProvider aclmgmt.ACLProvider
}

var hbmpcscclogger = flogging.MustGetLogger("hbmpcscc")

// These are function names from Invoke first parameter
const (
	//Communicate string = "Communicate"
	GetChainInfo       string = "GetChainInfo"
	GetBlockByNumber   string = "GetBlockByNumber"
	GetBlockByHash     string = "GetBlockByHash"
	GetTransactionByID string = "GetTransactionByID"
	GetBlockByTxID     string = "GetBlockByTxID"
)

// Init is called once per chain when the chain is created.
// This allows the chaincode to initialize any variables on the ledger prior
// to any transaction execution on the chain.
func (e *MpcEngine) Init(stub shim.ChaincodeStubInterface) pb.Response {
	hbmpcscclogger.Info("Init HBMPCSCC")

	return shim.Success(nil)
}

// Invoke is called with args[0] contains the query function name, args[1]
// contains the chain ID, which is temporary for now until it is part of stub.
// Each function requires additional parameters as described below:
// # GetChainInfo: Return a BlockchainInfo object marshalled in bytes
// # GetBlockByNumber: Return the block specified by block number in args[2]
// # GetBlockByHash: Return the block specified by block hash in args[2]
// # GetTransactionByID: Return the transaction specified by ID in args[2]
func (e *MpcEngine) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetArgs()

	if len(args) < 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments, %d", len(args)))
	}
	fname := string(args[0])
	cid := string(args[1])

	if fname != GetChainInfo && len(args) < 3 {
		return shim.Error(fmt.Sprintf("missing 3rd argument for %s", fname))
	}

	targetLedger := peer.GetLedger(cid)
	if targetLedger == nil {
		return shim.Error(fmt.Sprintf("Invalid chain ID, %s", cid))
	}

	hbmpcscclogger.Debugf("Invoke function: %s on chain: %s", fname, cid)

	// Handle ACL:
	// 1. get the signed proposal
	sp, err := stub.GetSignedProposal()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed getting signed proposal from stub, %s: %s", cid, err))
	}

	// 2. check the channel reader policy
	res := getACLResource(fname)
	if err = e.aclProvider.CheckACL(res, cid, sp); err != nil {
		return shim.Error(fmt.Sprintf("access denied for [%s][%s]: [%s]", fname, cid, err))
	}

	switch fname {
	case GetTransactionByID:
		return getTransactionByID(targetLedger, args[2])
	case GetBlockByNumber:
		return getBlockByNumber(targetLedger, args[2])
	case GetBlockByHash:
		return getBlockByHash(targetLedger, args[2])
	case GetChainInfo:
		return getChainInfo(targetLedger)
	case GetBlockByTxID:
		return getBlockByTxID(targetLedger, args[2])
	}

	return shim.Error(fmt.Sprintf("Requested function %s not found.", fname))
}

func getTransactionByID(vledger ledger.PeerLedger, tid []byte) pb.Response {
	if tid == nil {
		return shim.Error("Transaction ID must not be nil.")
	}

	processedTran, err := vledger.GetTransactionByID(string(tid))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get transaction with id %s, error %s", string(tid), err))
	}

	bytes, err := utils.Marshal(processedTran)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func getBlockByNumber(vledger ledger.PeerLedger, number []byte) pb.Response {
	if number == nil {
		return shim.Error("Block number must not be nil.")
	}
	bnum, err := strconv.ParseUint(string(number), 10, 64)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to parse block number with error %s", err))
	}
	block, err := vledger.GetBlockByNumber(bnum)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get block number %d, error %s", bnum, err))
	}
	// TODO: consider trim block content before returning
	//  Specifically, trim transaction 'data' out of the transaction array Payloads
	//  This will preserve the transaction Payload header,
	//  and client can do GetTransactionByID() if they want the full transaction details

	bytes, err := utils.Marshal(block)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func getBlockByHash(vledger ledger.PeerLedger, hash []byte) pb.Response {
	if hash == nil {
		return shim.Error("Block hash must not be nil.")
	}
	block, err := vledger.GetBlockByHash(hash)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get block hash %s, error %s", string(hash), err))
	}
	// TODO: consider trim block content before returning
	//  Specifically, trim transaction 'data' out of the transaction array Payloads
	//  This will preserve the transaction Payload header,
	//  and client can do GetTransactionByID() if they want the full transaction details

	bytes, err := utils.Marshal(block)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func getChainInfo(vledger ledger.PeerLedger) pb.Response {
	binfo, err := vledger.GetBlockchainInfo()
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get block info with error %s", err))
	}
	bytes, err := utils.Marshal(binfo)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func getBlockByTxID(vledger ledger.PeerLedger, rawTxID []byte) pb.Response {
	txID := string(rawTxID)
	block, err := vledger.GetBlockByTxID(txID)

	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get block for txID %s, error %s", txID, err))
	}

	bytes, err := utils.Marshal(block)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(bytes)
}

func getACLResource(fname string) string {
	return "hbmpcscc/" + fname
}
