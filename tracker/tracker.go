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
func Verify_join_callback(cbft_address string, trackerIns *Tracker) (determination bool) {
	for pass := range trackerIns.ValidatorList {
		if trackerIns.ValidatorList[pass].validatorAddress == cbft_address {
			// Add this tokenId to the join tx. This callback verifies that you have redeemed correctly.
			// Could return token id here if we need to.
			return true
		}
	}
	return false
}

func Verify_ValidatorRedeemEvent_callback(cbft_address string, tokenId string, trackerIns *Tracker) (determination bool) {
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

// The tracker will keep a list of validator pass redeem events.
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
func (nft_tracker *Tracker) StartTracking(ctx context.Context, contractAddress string, startBlock int, interval time.Duration, confirmations int) { // To-Do: Return an error channel

	// Initialise the ethereum RPC client
	ethereum_client, err := ethclient.Dial(nft_tracker.RpcAddress)
	if err != nil {
		panic(err) // Should surface an error about the rpc source if possible
	}
	defer ethereum_client.Close()
	// Record the latest block and search back in time to the start block until all historical redeem events are recorded.
	// Get the current block number
	latestBlock, noLatestBlock := ethereum_client.BlockNumber(ctx)
	if noLatestBlock != nil {
		panic(noLatestBlock)
	}
	nft_tracker.FetchHistoricalRedeems(nft_tracker.RpcAddress, contractAddress, startBlock, int(latestBlock))
	if err != nil {
		panic(err)
	}
	latestCheckedBlock := int(latestBlock)

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-nft_tracker.quit:
			fmt.Println("Received quit signal, stopping...")
			return // Exit the goroutine
		case <-ticker.C:
			// Subscribe to event stream and update accordingly
			latestBlock, noLatestBlock := ethereum_client.BlockNumber(ctx)
			if noLatestBlock != nil {
				panic(noLatestBlock)
			}
			if int(latestBlock)-confirmations > latestCheckedBlock {
				// Find all redeem events from deployblock to latest block.
				nft_tracker.FetchHistoricalRedeems(nft_tracker.RpcAddress, contractAddress, latestCheckedBlock, int(latestBlock))
				if err != nil {
					panic(err)
				}
				latestCheckedBlock = int(latestBlock)
			} else {
				fmt.Println("No new blocks searched since interval. Latest block is:", latestBlock, "  while last checked block was: ", latestCheckedBlock)
			}
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
func (nft_tracker *Tracker) FetchHistoricalRedeems(rpcSource string, contractAddress string, fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
	RedeemsFound := []Validator_RedeemEvent{}
	lastUpdate := 0
	maxBlockSearch := 4 // Number of blocks to increment by at a time.
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
				RedeemsFound = append(RedeemsFound, ValidatorList[vpass])
				nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])
			}
		}
		ProgressUpdate(fromBlock, currentBlock, toBlock, &lastUpdate)

	}
	return RedeemsFound, nil
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
			"0x4fc9c25b46f7854a495f8830e3d532a48cd64b4e4e3f6038557fe5669885bbe6", // Keccak256 hash, is the event signature of: Redeemed(uint256,bytes32)
		},
	}

	// fmt.Printf("Searching for redeem event logs in blocks %d to %d\n", fromBlock, toBlock)
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
		//fmt.Println("No response to rpc.")
	}

	return redeemEventsInRange, nil
}

// Test function
func ProgressUpdate(startBlock int, currentBlock int, toBlock int, lastUpdate *int) {
	currentBlockPosition := currentBlock - startBlock
	destinationBlockPosition := toBlock - startBlock
	progress := (float64(currentBlockPosition) / float64(destinationBlockPosition)) * 100

	if int(progress) > *lastUpdate { // Only update every 1% completion
		fmt.Printf("Completion percentage of search: %d%%\n", int(progress))
		*lastUpdate = int(progress)
		//fmt.Printf("Searching for redeem events in blocks %d to %d\n", currentBlock, currentBlock+4)
	}
}
