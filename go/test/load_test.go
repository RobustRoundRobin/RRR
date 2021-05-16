// build quorumloadtest

package test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var NODE_ENDPOINT = "http://127.0.0.1:8300"

// var TESSERA_ENDPOINT = "http://127.0.0.1:9008"
var TESSERA_ENDPOINT = ""

// Set to roughly the configured round length
var expectedLatency = time.Second * 6

var failedRoundTolerance = 3

var getSetAddABI = `[ { "constant": false, "inputs": [ { "internalType": "uint256", "name": "x", "type": "uint256" } ], "name": "add", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": false, "inputs": [ { "internalType": "uint256", "name": "x", "type": "uint256" } ], "name": "set", "outputs": [], "payable": false, "stateMutability": "nonpayable", "type": "function" }, { "constant": true, "inputs": [], "name": "get", "outputs": [ { "internalType": "uint256", "name": "retVal", "type": "uint256" } ], "payable": false, "stateMutability": "view", "type": "function" } ]`
var getSetAddBin = "0x608060405234801561001057600080fd5b50610126806100206000396000f3fe6080604052348015600f57600080fd5b50600436106059576000357c0100000000000000000000000000000000000000000000000000000000900480631003e2d214605e57806360fe47b11460895780636d4ce63c1460b4575b600080fd5b608760048036036020811015607257600080fd5b810190808035906020019092919050505060d0565b005b60b260048036036020811015609d57600080fd5b810190808035906020019092919050505060de565b005b60ba60e8565b6040518082815260200191505060405180910390f35b806000540160008190555050565b8060008190555050565b6000805490509056fea265627a7a72315820c8bd9d7613946c0a0455d5dcd9528916cebe6d6599909a4b2527a8252b40d20564736f6c634300050b0032"

// var getSetAddAddrHex = "0x289aC6896B7AD90d5363bd84584AEd7D420cb2C5"
// var getSetAddPublicAddrHex = "0x3e4f15b878bdc3484082d3340a76922952945d55"
var defaultGasLimit = uint64(60000)

// TODOD: fetch these from the 3rdparty-api.
var privateFor = []string{
	//	"i7zKPttxhdAjYMLAtrwvLAL60o6tVJ46OBALBe+JBgE=", // dev-robin-0
	"xD4CVnwaCc+7C8yDEgQVHED8rd4K25laO/E5S460ehs=", // dev-robin-1
	"Y/6KKt1oQpuF06ajIEbc/yahB+Fp9kDNtUqdb5oGhWc=", // dev-robin-2
}

var wallets = []string{
	"0x0875df9482284bc516023fc93fda79a2e5632675",
	"0x9f00fa6d57f790939debba93bd7206153b1951bc",
	"0x1a5665b682df1dbfd6c52bbaaee5b024ee9c391a",
	"0x988bf41e5bf7c3e6fa13ce18ba1e0ad70f947c0f",
	"0x0ec36c04031b238bb935c8354a4ed6e5d786e588",
	"0x3388c9f74849355c4b4cc000165078c03d27d8dd",
	"0x9c2a5cfd692dbebb701151d926831db4def30937",
	"0x240a9d756ce58114aa656371dd883edfbe7151e9",
	"0xb44de397cbdfb2f2636395cc677b07561967439f",
	"0x6dec5dc3518dcebb9f2cc70c5915bc7a6fc64b72",
	"0xf89e2361fb5b2eeb7229aa1f78e75a3955bf247f",
	"0xa92c5dcc151cbf967a1b102f81872c323bf0d79e",
	"0x0a52cb84d427562afa9df1cebeeb21c61e0cebe8",
	"0xb87b094f9cfc4067c21da5387cc21ee06aa72152",
	"0x7a724e6c760314432b8c18496bf2a1cbfed7eed1",
	"0xba2c9f827b343b8c98362e2cc91619ea8dbffcda",
	"0x593526dc52f7f1666b0711371b8cb37551a1e2e8",
	"0xf7584eb30123725a7a55ac77a0f265291d3685be",
	"0x6d2448839d9578db309832f82d4c20e88937a6b4",
	"0xc10b8ea2c9951632a352a1edbcb786e8903a9e05",
	"0x785d2ccee00c7715cd845f0a33bdcb9ae5cf5bb3",
}

var keys = []string{
	"6a21b55d1903e1d03fe8d5d9f4e1e1d848939a357d6a12d4ad00bf25c74f6367",
	"6cb13982ddfbf60849fc9971ad532ffa74f58e401f86b5a0c21e0bd79446dcb5",
	"77b8c8d261dc077c500da922b60ecbadae7bedee7fe43edd2d7db539b8fa45b7",
	"4055ce137bfe0803434a5c4d2d5bbc7c5eea848ee8d9b413502df09dbb9f0e37",
	"8c27eea5933984ae51e130c3c7b8a05ae37597aab6e40ca3f8de350927b98022",
	"28b9042f541ba34994eb88789d84f65a45433fc744f7150d5bdddd19a6664692",
	"928ffc4fc87932622eda240919c9c136750304742161e237cf111d3174f0907e",
	"bbb734c7fbdb7b9e9251b098723a36c02ead1e879058c3b0eaef1516be93255b",
	"8221b7c736e4e52adb047d230220ab2b71cc8e54d4bd9333801d979a0dc8f616",
	"ad6bf68caafdb01222192e56f04dec3b85201909bc31e1b499bb75e460d50899",
	"b8cc242d0fdbb4cc921a72585aa3da5752518e5f6a171baee9a3235eb6403c1e",
	"cf14d1e182170d669977002353753aee0bc5af50a68e1412ed02cd1d014fbed2",
	"a26736b22fc395a9fc1144ac6d2354f9014912cb8e2b41a00a7101ea1de6e87d",
	"f8aa555ef965739048e915d117d2889685c58c2c238879870cfcf0b2b12d5abf",
	"0e6d7a74423d0f94e50c7da18ccb57701bd36596c78a5d375167fb376f27a6ec",
	"689e6c2d94ad2a26c6a07b31549e65d12a1618bf2384547bf8bf14986f6d8124",
	"f189a48ad5baf6117dd23e3ec7890b4d4eb78e8b67a5db130808fc72ac921543",
	"79cb5e4129002130655844cf3b9e4b4315c4589ccfdefd7f3d5cf60dc50a315b",
	"3fefc3b30b5572ed26d312e31dc97f502442d39199ddb497c4f350d832c5e7c0",
	"102e10988f2384c7f39b006fa49e628c40356ebfc108de80f2ddb52a307a01ad",
	"cfc994b98e1ed2d91e6d2c653efd7452e58ea4edd5749a5fbf6faa576a7e1d95",
}

var big1 = big.NewInt(1)

type QuorumSuite struct {
	suite.Suite
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

	s.public = true

	s.assert = assert.New(s.T())
	s.require = require.New(s.T())
	s.require.True(len(wallets) == len(keys))
	s.numAccounts = len(wallets)
	s.wallets = make([]common.Address, s.numAccounts)
	for i, hexAddr := range wallets {
		s.wallets[i] = common.HexToAddress(hexAddr)
	}
	s.keys = make([]*ecdsa.PrivateKey, s.numAccounts)
	s.auth = make([]*bind.TransactOpts, s.numAccounts)
	s.nonce = make([]uint64, s.numAccounts)
	for i, hexKey := range keys {
		var err error
		s.keys[i], err = crypto.HexToECDSA(hexKey)
		s.require.NoError(err)
		s.auth[i] = bind.NewKeyedTransactor(s.keys[i])
		if !s.public {
			s.auth[i].PrivateFor = make([]string, len(privateFor))
			copy(s.auth[i].PrivateFor, privateFor)
		}
		s.auth[i].GasLimit = uint64(defaultGasLimit)
		s.auth[i].GasPrice = big.NewInt(0)
	}

	ethC, err := newTransactor(NODE_ENDPOINT, TESSERA_ENDPOINT)
	parsed, err := abi.JSON(strings.NewReader(getSetAddABI))
	var tx *types.Transaction

	s.auth[0].GasLimit = uint64(500000000)
	s.address, tx, s.getSetAdd, err = bind.DeployContract(
		s.auth[0], parsed, common.FromHex(getSetAddBin), ethC)
	s.require.NoError(err)
	s.require.True(checkReceipt(ethC, tx))
	s.auth[0].GasLimit = uint64(defaultGasLimit)
}

func (s *QuorumSuite) TearDownSuite() {
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

func newBoundContract(address common.Address, abiJSON string, ethC *ethclient.Client) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, ethC, ethC, ethC), nil
}

func checkReceipt(ethC *ethclient.Client, tx *types.Transaction) bool {

	attempts := failedRoundTolerance * 3 // three attempts per expected round duration

	interval := expectedLatency / 3

	for i := 0; i < attempts; i++ {
		time.Sleep(interval)
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		r, err := ethC.TransactionReceipt(ctx, tx.Hash())
		cancel()
		if r == nil || err != nil {
			continue
		}
		if r.Status == 1 {
			return true
		}
		return false
	}
	return false
}

func adder(s *QuorumSuite, wg *sync.WaitGroup, first, last, rounds int, ethEndpoint, tessEndpoint string, check bool) {

	defer wg.Done()

	ethC, err := newTransactor(ethEndpoint, tessEndpoint)
	s.require.NoError(err)
	var tx *types.Transaction
	for r := 0; r < rounds; r++ {
		for i := first; i < last; i++ {
			tx, err = s.getSetAdd.Transact(s.auth[i], "add", big.NewInt(3))
			s.require.NoError(err)
			s.auth[i].Nonce.Add(s.auth[i].Nonce, big1)
			if check {
				ok := checkReceipt(ethC, tx)
				if !ok {
					fmt.Println("!ok")
				}
				s.require.True(ok)
			}
		}
	}
}

func (s *QuorumSuite) TestOneTransact() {
	ethC, err := newTransactor(NODE_ENDPOINT, TESSERA_ENDPOINT)
	defer ethC.Close()

	var nonce uint64
	nonce, err = ethC.PendingNonceAt(context.Background(), s.wallets[0])
	s.require.NoError(err)
	s.auth[0].Nonce = big.NewInt(int64(nonce))
	tx, err := s.getSetAdd.Transact(s.auth[0], "add", big.NewInt(3))
	s.require.True(checkReceipt(ethC, tx))
}

func (s *QuorumSuite) TestQuorum() {

	ethC, err := newTransactor(NODE_ENDPOINT, TESSERA_ENDPOINT)
	defer ethC.Close()

	// Initialise once and assume each go rountine manages its own entries
	for i := 0; i < s.numAccounts; i++ {
		var nonce uint64
		nonce, err = ethC.PendingNonceAt(context.Background(), s.wallets[i])
		s.require.NoError(err)
		s.auth[i].Nonce = big.NewInt(int64(nonce))
	}

	THREADS := 20
	BATCH_LEN := 1
	BATCH_PER_THREAD := 100
	var wg sync.WaitGroup
	for i := 0; i < THREADS; i++ {
		// if we have multiple exposed nodes we can do this
		// ethEndpoint := fmt.Sprintf("http://localhost:220%02d", i)
		// tessEndpoint := fmt.Sprintf("http://localhost:90%02d", 8+i)
		// Otherwise we hit the same node with multiple clients
		ethEndpoint := NODE_ENDPOINT
		tessEndpoint := TESSERA_ENDPOINT

		wg.Add(1)
		go adder(s, &wg, i, i+BATCH_LEN, BATCH_PER_THREAD, ethEndpoint, tessEndpoint, false)
	}
	wg.Wait()
	ntx := THREADS * BATCH_LEN * BATCH_PER_THREAD
	fmt.Printf("completed: %d\n", ntx)

}
func TestQuorum(t *testing.T) {
	suite.Run(t, new(QuorumSuite))
}
