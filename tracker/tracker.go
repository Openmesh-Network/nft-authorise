package validatorpass_tracker

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// COMETBFT CALLBACKS

// CometBFT callback without requiring tokenId, to determine validity of cometbft address in terms of existence of an on-chain redeem event.
// This function should be called before the validator tries to initiate a join transaction to the network.
func VerifyMembershipOfAddress(cometBftAddress string, trackerIns *Tracker) (determination bool) {
	for pass := range trackerIns.ValidatorList {
		if trackerIns.ValidatorList[pass].validatorAddress == cometBftAddress {
			return true
		}
	}
	return false
}

// CometBFT callback to determine validity of cometbft address in terms of existence of an on-chain redeem event.
func VerifyValidatorAddress(cometBftAddress string, tokenId string, trackerIns *Tracker) (determination bool) {
	EventsForTokenId := []Validator_RedeemEvent{}
	for pass := range trackerIns.ValidatorList {
		if trackerIns.ValidatorList[pass].tokenId == tokenId {
			if trackerIns.ValidatorList[pass].validatorAddress == cometBftAddress {
				EventsForTokenId = append(EventsForTokenId, trackerIns.ValidatorList[pass]) // Get all events for this tokenId
			}
		}
	}
	searchHeight := 0
	latestEvent := Validator_RedeemEvent{}
	for pass := range EventsForTokenId {
		if EventsForTokenId[pass].redeemedBlockHeight > searchHeight {
			latestEvent = EventsForTokenId[pass]
			searchHeight = EventsForTokenId[pass].redeemedBlockHeight // Find the latest event
		}
	}
	if latestEvent.validatorAddress == cometBftAddress {
		return true
	} else {
		return false
	}
}

// The tracker will keep a list of validator pass redeem events.
type Tracker struct {
	RpcAddress        string
	rpcSearchLimit    int
	TrackedEvent      Rpc_RedeemEvent
	ValidatorList     []Validator_RedeemEvent
	lastTrackerHeight int
}

// Create a new tracker object to track an event.
func NewTracker(rpcSourceAddress string, rpcSearchLimit int, TrackedEvent Rpc_RedeemEvent) *Tracker {
	return &Tracker{
		RpcAddress:        rpcSourceAddress,
		rpcSearchLimit:    rpcSearchLimit,
		TrackedEvent:      TrackedEvent,
		ValidatorList:     []Validator_RedeemEvent{},
		lastTrackerHeight: 0,
	}
}

// Start tracking redeem events from a Validator Pass smart contract address, you should be able to deterministically call validateNFTMembership()
// for peer validation in a CometBFT callback.
func (nft_tracker *Tracker) StartTracking(ctx context.Context, interval time.Duration, confirmations int) (errChannel chan error) { // To-Do: Error channel logic
	ethereum_client, err := ethclient.Dial(nft_tracker.RpcAddress)
	if err != nil {
		panic(err) // Crash if no access to RPC.
	}
	defer ethereum_client.Close()
	// Record the latest block and search back in time to the start block until all historical redeem events are recorded.
	// Get the current block number
	latestBlock, noLatestBlock := ethereum_client.BlockNumber(ctx)
	if noLatestBlock != nil {
		errChannel <- noLatestBlock
	}
	// Do a historical search of all redeem events from deployBlock to latestBlock
	list, err := nft_tracker.FindRedeems(nft_tracker.TrackedEvent.deployBlock, int(latestBlock))
	if err != nil {
		errChannel <- noLatestBlock
	}
	fmt.Println("Found", len(list), "redeem events in historical search")
	latestCheckedBlock := int(latestBlock)

	ticker := time.NewTicker(interval)
	for range ticker.C {

		// Subscribe to event stream and update accordingly
		latestBlock, noLatestBlock := ethereum_client.BlockNumber(ctx)
		if noLatestBlock != nil {
			panic(noLatestBlock) // Need to investigate potential errors that could be surfaced here.
		}
		elgibleBlock := int(latestBlock) - confirmations // Block eligible to be searched based on confirmation parameter
		if elgibleBlock > latestCheckedBlock {
			// Find all redeem events from deployblock to latest block.
			nft_tracker.FindRedeems(latestCheckedBlock, elgibleBlock)
			latestCheckedBlock = int(latestBlock)
		} else {
			fmt.Println("No new blocks searched since interval. Latest block is:", latestBlock, "  while last checked block was: ", latestCheckedBlock)
		}

	}
	return errChannel
}

// RPC FUNCTIONS

// Used to make many ethereum remote procedure calls over time to handle limits from rpc provider.
// Setting a maxBlockSearch of 0 will assume that you have unlimited RPC access, eg. lite or full node locally hosted.
func (nft_tracker *Tracker) FindRedeems(fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
	RedeemsFound := []Validator_RedeemEvent{}
	lastUpdate := 0
	if nft_tracker.rpcSearchLimit == 0 { // Unlimited RPC, no need to search incrementally.
		nft_tracker.FetchAppendRedeems(nft_tracker.TrackedEvent.deployBlock, toBlock)
	} else {
		for currentBlock := fromBlock; currentBlock <= toBlock; currentBlock += nft_tracker.rpcSearchLimit + 1 {
			// Search through all blocks incrementing by rpcSearchLimit.
			nft_tracker.FetchAppendRedeems(currentBlock, (currentBlock + nft_tracker.rpcSearchLimit))
			ProgressUpdate(fromBlock, currentBlock, toBlock, &lastUpdate)

		}
	}

	return RedeemsFound, nil
}

func (nft_tracker *Tracker) FetchAppendRedeems(fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
	RedeemsFound := []Validator_RedeemEvent{}
	ValidatorList, err := FetchRedeemEventsRPC(nft_tracker.RpcAddress, nft_tracker.TrackedEvent, fromBlock, toBlock)
	if err != nil {
		return nil, err
	}
	if len(ValidatorList) > 0 {
		for vpass := range ValidatorList {
			fmt.Println(ValidatorList[vpass].validatorAddress) // For debugging

			RedeemsFound = append(RedeemsFound, ValidatorList[vpass])
			nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])
		}
		// Update nft_tracker.lastTrackerHeight
		nft_tracker.lastTrackerHeight = toBlock
	}
	return RedeemsFound, nil
}

// Fetch a full list of Validator Passes from a smart contract address.
func FetchRedeemEventsRPC(rpcSource string, TrackedEvent Rpc_RedeemEvent, fromBlock int, toBlock int) ([]Validator_RedeemEvent, error) {
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
		"address":   TrackedEvent.contractAddress,
		"topics": []string{
			TrackedEvent.EventSignature, // Keccak256 hash, is the event signature of: Redeemed(uint256,bytes32)
		},
	}

	// fmt.Printf("Searching for redeem event logs in blocks %d to %d\n", fromBlock, toBlock)
	response := make([]RedeemEventRpc, 1)
	// To-Do: Handle responses that are error messages, for example if the RPC is down.
	getLogsFailed := ethereum_client.Client().Call(&response, "eth_getLogs", RpcArguments)
	if getLogsFailed != nil {
		panic(getLogsFailed)
	}

	// response handling logic
	if len(response) > 0 {
		redeemEventsInRange = append(redeemEventsInRange, *NewValidatorRedeemEvent(response[0].Topics[1], response[0].Data)) // Hard coded interaction with JSON response.
	}

	return redeemEventsInRange, nil
}

// Verbose function for printing the progress of the search
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
