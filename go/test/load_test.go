// build quorumloadtest

// This tool is most conveniently used by "adding a workspace folder" in visual
// studio code. The directory containing the go.mod file is the folder to add.
// With that done, vscode can be used to install and configure the go tooling
// necessary to run this tool. Treat it like a 12 factor app and use settings
// and launch.jsons to configure the test runs. If everything is setup
// correctly in vscode, TestQuorum will have a little grey "run test" hyperlink
// above it.

package test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var getSetAddABI = `[ { "constant": false, "inputs": [ { "internalType": "uint256", "name": "x", "type": "uint256" } ], "name": "add", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": false, "inputs": [ { "internalType": "uint256", "name": "x", "type": "uint256" } ], "name": "set", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "get", "outputs": [ { "internalType": "uint256", "name": "retVal", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" } ]`
var getSetAddBin = "0x608060405234801561001057600080fd5b50610126806100206000396000f3fe6080604052348015600f57600080fd5b50600436106059576000357c0100000000000000000000000000000000000000000000000000000000900480631003e2d214605e57806360fe47b11460895780636d4ce63c1460b4575b600080fd5b608760048036036020811015607257600080fd5b810190808035906020019092919050505060d0565b005b60b260048036036020811015609d57600080fd5b810190808035906020019092919050505060de565b005b60ba60e8565b6040518082815260200191505060405180910390f35b806000540160008190555050565b8060008190555050565b6000805490509056fea265627a7a72315820c8bd9d7613946c0a0455d5dcd9528916cebe6d6599909a4b2527a8252b40d20564736f6c634300050b0032"

var big0 = big.NewInt(0)
var big1 = big.NewInt(1)

func TestQuorum(t *testing.T) {
	suite.Run(t, new(QuorumSuite))
}

type QuorumSuite struct {
	suite.Suite

	// All the CAPS variables are config vars. linter can get lost.
	NODE_ENDPOINT    string
	TESSERA_ENDPOINT string
	RECEIPT_RETRIES  int
	EXPECTED_LATENCY time.Duration
	DEFAULT_GASLIMIT uint64
	PRIVATE_FOR      []string // "base64pub:base64pub:..." defaults empty - no private tx

	// If true, confirm every transaction in a batch before doing the next batch.
	CHECK_BATCH_RECEIPTS bool

	DEPLOY_KEY string   // needs to have funds even for quorum, used to deploy contract
	FROM_KEYS  []string // 0 or more : seperated hex private keys

	NUM_GENERATED_FROM_KEYS int // number of private keys to generate

	// We distribute the accounts evenly accross the threads so that Each thread
	// issues transactions for [len(FROM_KEYS) + NUM_GENERATED_FROM_KEYS] / NUM_THREADS
	NUM_THREADS int

	// The maximum number of nodes to spread the transactions over. We assume we
	// can simply add to the base port in NODE_ENDPOINT for this many nodes
	MAX_NODES int

	// Each thread will issue at most NUM_TRANSACTIONS / NUM_THREADS
	NUM_TRANSACTIONS int

	// batch len = NUM_FROM_KEYS / NUM_THREADS
	// batch per thread = NUM_TRANSACTIONS / NUM_THREADS

	assert      *assert.Assertions
	require     *require.Assertions
	numAccounts int
	wallets     []common.Address
	keys        []*ecdsa.PrivateKey
	auth        []*bind.TransactOpts
	nonce       []uint64
	public      bool
	getSetAdd   *bind.BoundContract
	address     common.Address
}

func (s *QuorumSuite) SetupSuite() {

	s.NODE_ENDPOINT = fromEnv("NODE_ENDPOINT", "http://127.0.0.1:8545")
	s.TESSERA_ENDPOINT = fromEnv("NODE_ENDPOINT", "http://127.0.0.1:9008")
	s.EXPECTED_LATENCY = durationFromEnv("EXPECTED_LATENCY", 3*time.Second)
	s.RECEIPT_RETRIES = intFromEnv("RECEIPT_RETRIES", 15)
	s.DEFAULT_GASLIMIT = uint64(intFromEnv("DEFAULT_GASLIMIT", 6000000))
	s.CHECK_BATCH_RECEIPTS = intFromEnv("CHECK_BATCH_RECIEPTS", 0) > 0

	privateFor := fromEnv("PRIVATE_FOR", "")
	if privateFor != "" {
		s.PRIVATE_FOR = strings.Split(privateFor, ":")
	}
	s.DEPLOY_KEY = fromEnv("DEPLOY_KEY", "")
	s.NUM_GENERATED_FROM_KEYS = intFromEnv("NUM_GENERATED_FROM_KEYS", 200)
	fromKeys := fromEnv("FROM_KEYS", "")
	if fromKeys != "" {
		s.FROM_KEYS = strings.Split(fromKeys, ":")
	}
	s.NUM_THREADS = intFromEnv("NUM_THREADS", 20)
	s.MAX_NODES = intFromEnv("MAX_NODES", 1)
	s.NUM_TRANSACTIONS = intFromEnv("NUM_TRANSACTIONS", 2000)

	s.public = true

	s.assert = assert.New(s.T())
	s.require = require.New(s.T())

	s.numAccounts = s.NUM_GENERATED_FROM_KEYS + len(s.FROM_KEYS)

	s.wallets = make([]common.Address, s.numAccounts)
	s.keys = make([]*ecdsa.PrivateKey, s.numAccounts)
	s.auth = make([]*bind.TransactOpts, s.numAccounts)
	s.nonce = make([]uint64, s.numAccounts)
	for i, hexKey := range s.FROM_KEYS {
		var err error
		s.keys[i], err = crypto.HexToECDSA(hexKey)
		s.require.NoError(err)

		pub := s.keys[i].PublicKey

		// derive the wallet address from the private key
		pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
		pubHash := crypto.Keccak256(pubBytes[1:]) // skip the compression indicator
		copy(s.wallets[i][:], pubHash[12:])       // wallet address is the trailing 20 bytes

		s.auth[i] = bind.NewKeyedTransactor(s.keys[i])
		if !s.public {
			s.auth[i].PrivateFor = make([]string, len(s.PRIVATE_FOR))
			copy(s.auth[i].PrivateFor, s.PRIVATE_FOR)
		}
		s.auth[i].GasLimit = s.DEFAULT_GASLIMIT
		s.auth[i].GasPrice = big.NewInt(0)
	}

	for j := 0; j < s.NUM_GENERATED_FROM_KEYS; j++ {
		ik := len(s.FROM_KEYS) + j

		var err error
		s.keys[ik], err = crypto.GenerateKey()
		s.require.NoError(err)

		pub := s.keys[ik].PublicKey

		// derive the wallet address from the private key
		pubBytes := elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
		pubHash := crypto.Keccak256(pubBytes[1:]) // skip the compression indicator
		copy(s.wallets[ik][:], pubHash[12:])      // wallet address is the trailing 20 bytes

		s.auth[ik] = bind.NewKeyedTransactor(s.keys[ik])
		if !s.public {
			s.auth[ik].PrivateFor = make([]string, len(s.PRIVATE_FOR))
			copy(s.auth[ik].PrivateFor, s.PRIVATE_FOR)
		}
		s.auth[ik].GasLimit = s.DEFAULT_GASLIMIT
		s.auth[ik].GasPrice = big.NewInt(0)
	}

	ethC, err := newTransactor(s.NODE_ENDPOINT, s.TESSERA_ENDPOINT)
	s.require.NoError(err)
	parsed, err := abi.JSON(strings.NewReader(getSetAddABI))
	s.require.NoError(err)

	var tx *types.Transaction

	var deployKey *ecdsa.PrivateKey
	if s.DEPLOY_KEY != "" {
		deployKey, err = crypto.HexToECDSA(s.DEPLOY_KEY)
		s.require.NoError(err)
	}
	if deployKey == nil {
		// This will likely fail as normal quorum requires balance to deploy
		// event tho gasprice is 0
		deployKey, err = crypto.GenerateKey()
		s.require.NoError(err)
	}

	deployAuth := bind.NewKeyedTransactor(deployKey)

	deployAuth.GasLimit = uint64(500000000)
	deployAuth.GasPrice = big0
	s.address, tx, s.getSetAdd, err = bind.DeployContract(
		deployAuth, parsed, common.FromHex(getSetAddBin), ethC)
	s.require.NoError(err)
	s.require.True(s.checkReceipt(ethC, tx))
}

func (s *QuorumSuite) TearDownSuite() {
}

// TestOneTransact succeedes if a single "add" transaction can be made for the
// test contract.
func (s *QuorumSuite) TestOneTransact() {
	ethC, err := newTransactor(s.NODE_ENDPOINT, s.TESSERA_ENDPOINT)
	s.require.NoError(err)
	defer ethC.Close()

	var nonce uint64
	nonce, err = ethC.PendingNonceAt(context.Background(), s.wallets[0])
	s.require.NoError(err)
	s.auth[0].Nonce = big.NewInt(int64(nonce))
	tx, err := s.getSetAdd.Transact(s.auth[0], "add", big.NewInt(3))
	s.require.NoError(err)
	s.require.True(s.checkReceipt(ethC, tx))
}

func adder(
	s *QuorumSuite, wg *sync.WaitGroup, banner string, checkBatchReceipts bool, first, last, numBatches int, ethEndpoint, tessEndpoint string) {

	defer wg.Done()

	ethC, err := newTransactor(ethEndpoint, tessEndpoint)
	s.require.NoError(err)
	var tx *types.Transaction
	batch := make([]*types.Transaction, last-first)

	for r := 0; r < numBatches; r++ {
		fmt.Printf("%s: batch %d, eth %s, tess %s\n", banner, r, ethEndpoint, tessEndpoint)
		for i := first; i < last; i++ {
			tx, err = s.getSetAdd.Transact(s.auth[i], "add", big.NewInt(3))
			s.require.NoError(err)
			s.auth[i].Nonce.Add(s.auth[i].Nonce, big1)
			batch[i-first] = tx
		}
		if checkBatchReceipts {
			for i := first; i < last; i++ {
				ok := s.checkReceipt(ethC, batch[i-first])
				if !ok {
					fmt.Println("!ok")
				}
				s.require.True(ok)
			}
		}
	}
}

// TestQuorum issues "add" transactions from multiple threads. Note that it is
// not very chatty.
func (s *QuorumSuite) TestQuorum() {

	ethC, err := newTransactor(s.NODE_ENDPOINT, s.TESSERA_ENDPOINT)
	s.require.NoError(err)
	defer ethC.Close()

	// Initialise once and assume each go rountine manages its own entries
	for i := 0; i < s.numAccounts; i++ {
		var nonce uint64
		nonce, err = ethC.PendingNonceAt(context.Background(), s.wallets[i])
		s.require.NoError(err)
		s.auth[i].Nonce = big.NewInt(int64(nonce))
	}

	BATCH_LEN := s.numAccounts / s.NUM_THREADS
	s.require.NotEqual(BATCH_LEN, 0)
	BATCH_PER_THREAD := s.NUM_TRANSACTIONS / (s.NUM_THREADS * BATCH_LEN)

	qu, err := url.Parse(s.NODE_ENDPOINT)
	s.require.NoError(err)
	tu, err := url.Parse(s.TESSERA_ENDPOINT)
	s.require.NoError(err)

	quHostname := qu.Hostname()
	tuHostname := tu.Hostname()

	baseQuorumPort, err := strconv.Atoi(qu.Port())
	s.require.NoError(err)
	baseTesseraPort, err := strconv.Atoi(tu.Port())
	s.require.NoError(err)

	var wg sync.WaitGroup
	for i := 0; i < s.NUM_THREADS; i++ {

		// if we have multiple exposed nodes we can do this
		// ethEndpoint := fmt.Sprintf("http://localhost:220%02d", i)
		// tessEndpoint := fmt.Sprintf("http://localhost:90%02d", 8+i)
		// Otherwise we hit the same node with multiple clients

		qu.Host = fmt.Sprintf("%s:%d", quHostname, baseQuorumPort+(i%s.MAX_NODES))
		tu.Host = fmt.Sprintf("%s:%d", tuHostname, baseTesseraPort+(i%s.MAX_NODES))

		wg.Add(1)
		go adder(
			s, &wg, fmt.Sprintf("client-%d", i),
			s.CHECK_BATCH_RECEIPTS, i*BATCH_LEN, i*BATCH_LEN+BATCH_LEN, BATCH_PER_THREAD,
			qu.String(), tu.String())
	}
	wg.Wait()
	ntx := s.NUM_THREADS * BATCH_LEN * BATCH_PER_THREAD
	fmt.Printf("completed: %d\n", ntx)

}

func newTransactor(ethEndpoint, tesseraEndpoint string) (*ethclient.Client, error) {

	ethRPC, err := rpc.DialHTTPWithClient(ethEndpoint, &http.Client{Timeout: time.Second * 10})
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(ethRPC)
	if ethClient == nil {
		return nil, fmt.Errorf("failed creating ethclient")
	}

	if tesseraEndpoint != "" {
		ethClient, err = ethClient.WithPrivateTransactionManager(tesseraEndpoint)
		if err != nil {
			return nil, err
		}
	}
	return ethClient, nil
}

// derived from https://blog.gopheracademy.com/advent-2014/backoff/
var backoffms = []int{0, 500, 500, 3000, 3000, 5000, 5000, 8000, 8000, 10000, 10000}

func backoffDuration(nth int) time.Duration {
	if nth >= len(backoffms) {
		nth = len(backoffms) - 1
	}
	return time.Duration(jitter(backoffms[nth])) * time.Millisecond
}

func jitter(ms int) int {
	if ms == 0 {
		return 0
	}
	return ms/2 + rand.Intn(ms)
}

func (s *QuorumSuite) checkReceipt(ethC *ethclient.Client, tx *types.Transaction) bool {

	for i := 0; i < s.RECEIPT_RETRIES; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), s.EXPECTED_LATENCY)
		r, err := ethC.TransactionReceipt(ctx, tx.Hash())
		cancel()
		if r == nil || err != nil {
			// fmt.Printf("backoff & retry: err=%v\n", err)
			time.Sleep(backoffDuration(i))
			continue
		}
		if r.Status == 1 {
			return true
		}
		return false
	}
	return false
}

func fromEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func intFromEnv(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return value
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := time.ParseDuration(val)
	if err != nil {
		panic(err)
	}
	return value
}
