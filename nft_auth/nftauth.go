package nft_auth

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Validator_Pass struct {
	tokenId          string // NFT token ID
	validatorAddress string // CometBFT validator address eg. cometvaloper1abc123def456ghi789jkl123mno456pqr789stu
}

func NewValidatorPass(tokenId string, validatorAddress string) *Validator_Pass {
	return &Validator_Pass{
		tokenId:          tokenId,
		validatorAddress: validatorAddress,
	}
}

type Tracker struct {
	// Abstracts the source of Ethereum RPC to be used for NFT tracking. (Support for external or internal nodes)
	RpcAddress    string
	ValidatorList []Validator_Pass
	quit          chan struct{}
}

func NewTracker(rpcSourceAddress string) *Tracker {
	return &Tracker{
		RpcAddress:    rpcSourceAddress,
		ValidatorList: []Validator_Pass{},
		quit:          make(chan struct{}),
	}
}

// Start tracking redeem events from a Validator Pass smart contract address, you should be able to deterministically call validateNFTMembership() for peer validation in a CometBFT callback.
func (nft_tracker *Tracker) Start(contractAddress string, startBlock int) {
	nextStartBlock := startBlock

	// Initialise the ethereum RPC client
	ethereum_client, err := ethclient.Dial(nft_tracker.RpcAddress)
	if err != nil {
		panic(err) // Should surface an error about the rpc source if possible
	}
	defer ethereum_client.Close()

	//ticker := time.NewTicker(1 * time.Second)
	go func() {
		fmt.Println("Started goroutine")
		// First-time behaviour: record the latest block to ensure none are missed.

		for {
			select {
			case <-nft_tracker.quit:
				fmt.Println("Received quit signal, stopping...")
				return // Exit the goroutine
			default:
				// Fetch validator passes from start block to latest block (at the time of subscription)
				ValidatorList, err := FetchValidatorPassesRPC(nft_tracker.RpcAddress, contractAddress, string(nextStartBlock), string(nextStartBlock+5))
				if err != nil {
					panic(err)
				}

				// Update ValidatorList
				nft_tracker.ValidatorList = ValidatorList

				// Delay before fetching again.
				time.Sleep(1 * time.Second)
			}
		}
	}()

	for {
		select {
		case <-nft_tracker.quit:
			fmt.Println("Received quit signal, stopping...")
			return // Exit the goroutine
		default:
			// Subscribe to event stream and update accordingly
		}
	}

}
func (nft_tracker *Tracker) Stop() {
	// Stops the go routine in Start()
	close(nft_tracker.quit)
}

// Fetch a full list of Validator Passes from a smart contract address.
func FetchValidatorPassesRPC(rpcSource string, contractAddress string, fromBlock string, toBlock string) ([]Validator_Pass, error) {
	var validNFTs []Validator_Pass
	fmt.Println("Fetching validator passes from RPC")

	// Initialise an ethereum RPC client
	ethereum_client, err := ethclient.Dial(rpcSource)
	if err != nil {
		return nil, err
	}
	defer ethereum_client.Close()

	// Build the RPC arguments for eth_getLogs
	RpcArguments := map[string]interface{}{
		"fromBlock": fromBlock,
		"toBlock":   toBlock,
		"address":   "0x8D64aB58a17dA7d8788367549c513386f09a0A70",
		"topics": []string{
			"0x4fc9c25b46f7854a495f8830e3d532a48cd64b4e4e3f6038557fe5669885bbe6", // Keccak256 event signature of: Redeemed(uint256,bytes32)
		},
	}

	// Call eth_getLogs on RPC to find the any redeemed validator addresses.
	//ctxToPreventHanging, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//defer cancel()
	fmt.Println("Getting redeem event logs from RPC.")
	response := make([]RpcResult, 1)

	newerr := ethereum_client.Client().Call(&response, "eth_getLogs", RpcArguments)
	if newerr != nil {
		panic(newerr)
	}
	fmt.Println(response)

	// response handling logic
	validNFTs = append(validNFTs, *NewValidatorPass(response[0].Topics[1], response[0].Data))

	return validNFTs, nil
}

// Usage: if tokenId exists within the VPass list.
func (nft_tracker *Tracker) updateValidatorPass(ValPass Validator_Pass) {
	fmt.Println("Updating validator pass with tokenId:", ValPass.tokenId)
	for nft := range nft_tracker.ValidatorList {
		if nft_tracker.ValidatorList[nft].tokenId == ValPass.tokenId {
			nft_tracker.ValidatorList[nft].validatorAddress = ValPass.validatorAddress
		}
	}
}

type JsonRpc struct {
	// "jsonrpc":"2.0","id":1,"result":[]
	JsonVersion   string                 `json:"jsonrpc"`
	MsgIdentifier string                 `json:"id"`
	MsgResponse   map[string]interface{} //`json:"result"`
	/*struct {
		CometBftAddress string `json:"data"`
	} `json:"result"` */
}

type RpcResult struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}
