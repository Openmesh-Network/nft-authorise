package nft_auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// COMETBFT CALLBACKS

/*
Transaction callback - set this (cometBftAddress & tokenId) voting power to 1 for block proposer if their address is redeemed.
Verify callback - check this (CometBft address & tokenId) is redeemed under the new tokenId.
*/

// Consideration: Should this callback even need a tokenId as an input?
// Confirms a cbft address' existence in the Validator Pass redeem events. Is used by CometBFT ABCI before the program tries to issue a join tx.
func join_callback(cbft_address string, trackerIns *Tracker) (determination bool) {
	for pass := range trackerIns.ValidatorList {
		if trackerIns.ValidatorList[pass].validatorAddress == cbft_address {
			// Add this tokenId to the join tx. This callback verifies that you have redeemed correctly.
			// Could return token id here if we need to.
			return true
		}
	}
	return false
}

func verify_ValidatorRedeemEvent_callback(cbft_address string, tokenId string, trackerIns *Tracker) (determination bool) {
	for pass := range trackerIns.ValidatorList {
		if trackerIns.ValidatorList[pass].tokenId == tokenId {
			if trackerIns.ValidatorList[pass].validatorAddress == cbft_address {
				// Add this tokenId to the join tx. This callback verifies that you have redeemed correctly.
				return true
			}
		}
	}
	return false
}

// TRACKER
type Tracker struct {
	// Abstracts the source of Ethereum RPC to be used for NFT tracking. (Support for external or internal nodes)
	RpcAddress    string
	ValidatorList []Validator_RedeemEvent
	quit          chan struct{}
}

// Create a new tracker object to track Validator Passes.
func NewTracker(rpcSourceAddress string) *Tracker {
	return &Tracker{
		RpcAddress:    rpcSourceAddress,
		ValidatorList: []Validator_RedeemEvent{},
		quit:          make(chan struct{}),
	}
}

// Start tracking redeem events from a Validator Pass smart contract address, you should be able to deterministically call validateNFTMembership()
// for peer validation in a CometBFT callback.
func (nft_tracker *Tracker) StartTracking(ctx context.Context, contractAddress string, startBlock int, path any) { // To-Do: Return an error channel
	nextStartBlock := startBlock

	// Initialise the ethereum RPC client
	ethereum_client, err := ethclient.Dial(nft_tracker.RpcAddress)
	if err != nil {
		panic(err) // Should surface an error about the rpc source if possible
	}
	defer ethereum_client.Close()

	//ticker := time.NewTicker(1 * time.Second)
	go func() {
		fmt.Println("Started historical search goroutine")
		// Go routine purpose: record the latest block and search back in time to the start block until all historical redeem events are recorded.

		// Get the current block number
		latestBlock, noLatestBlock := ethereum_client.BlockNumber(ctx) // XXX: THis will block forever
		if noLatestBlock != nil {
			panic(noLatestBlock)
		}
		// Find all redeem events from deployblock to latest block.
		FetchHistoricalRedeems(nft_tracker.RpcAddress, contractAddress, startBlock, int(latestBlock))

		for {
			select {
			case <-nft_tracker.quit:
				fmt.Println("Received quit signal, stopping...")
				return // Exit the goroutine
			default:
				// Fetch validator passes from start block to latest block (at the time of subscription) MUMBO
				ValidatorList, err := FetchRedeemEventsRPC(nft_tracker.RpcAddress, contractAddress, nextStartBlock, nextStartBlock+2) // Returns validator pass list based on redeem events in the block range.
				if err != nil {
					panic(err)
				}
				// To-Do: Append not update. Should finish when all between startBlock & recorded latestBlock are responded to by RPC.
				for vpass := range ValidatorList {
					nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])
				}
				// To-Do: handle the different possible RPC responses including ones made by invalid requests or rate-limiting.

				// Update ValidatorList
				nft_tracker.ValidatorList = ValidatorList

				// Delay before fetching again.
				time.Sleep(1 * time.Second)
				nextStartBlock += 5
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
			// ethereum_client.Client().Subscribe()
		}
	}

}
func (nft_tracker *Tracker) Stop() {
	// Stops the go routine in Start()
	close(nft_tracker.quit)
}

// RPC FUNCTIONS

// Call (preferably in a go routine) ethereum RPC many times over a long period to fill in the validator list with all historical redeems.
// This function is only useful if there is a limit on block range per rpc request.
func FetchHistoricalRedeems(rpcSource string, contractAddress string, fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
	RedeemList := []Validator_RedeemEvent{}

	maxBlockSearch := 3 // Number of blocks to increment by at a time.
	//for currentStartBlock := range(fromBlock, toBlock)
	fmt.Println(maxBlockSearch)
	for currentBlock := fromBlock; currentBlock <= toBlock; currentBlock += maxBlockSearch + 1 {
		// Search through all blocks incrementing by maxBlockSearch.
		ValidatorList, err := FetchRedeemEventsRPC(rpcSource, contractAddress, currentBlock, currentBlock+maxBlockSearch)
		if err != nil {
			return nil, err
		}
		if len(ValidatorList) > 0 {
			for vpass := range ValidatorList {
				fmt.Println(ValidatorList[vpass].validatorAddress)
				RedeemList = append(RedeemList, ValidatorList[vpass])
			}
		}

	}
	return RedeemList, nil
}

// Fetch a full list of Validator Passes from a smart contract address.
func FetchRedeemEventsRPC(rpcSource string, contractAddress string, fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
	var redeemEventsInRange []Validator_RedeemEvent

	// Initialise an ethereum RPC client
	ethereum_client, err := ethclient.Dial(rpcSource)
	if err != nil {
		return nil, err
	}
	defer ethereum_client.Close()

	// Build the RPC arguments for eth_getLogs
	RpcArguments := map[string]interface{}{
		"fromBlock": string(fmt.Sprintf("%d", fromBlock)), // fromBlock,
		"toBlock":   string(fmt.Sprintf("%d", toBlock)),   // toBlock,
		"address":   contractAddress,
		"topics": []string{
			"0x4fc9c25b46f7854a495f8830e3d532a48cd64b4e4e3f6038557fe5669885bbe6", // Keccak256 event signature of: Redeemed(uint256,bytes32)
		},
	}

	// Call eth_getLogs on RPC to find the any redeemed validator addresses.
	//ctxToPreventHanging, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//defer cancel()
	fmt.Printf("Getting redeem event logs from blocks %d to %d\n", fromBlock, toBlock)
	response := make([]RedeemEventRpc, 1)
	newerr := ethereum_client.Client().Call(&response, "eth_getLogs", RpcArguments)
	if newerr != nil {
		panic(newerr)
	}

	// Check that the response returned

	// response handling logic
	if len(response) > 0 {
		fmt.Println(response[0].Data)
		redeemEventsInRange = append(redeemEventsInRange, *NewValidatorRedeemEvent(response[0].Topics[1], response[0].Data))
	} else {
		fmt.Println("No response to rpc.")
	}

	return redeemEventsInRange, nil
}
