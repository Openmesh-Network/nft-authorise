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

// Mapped search for cometBFT callback to account for re-redeems.
func VerifyAddress(cometBftAddress string, trackerIns *Tracker) bool {
	for address, redeems := range trackerIns.addressMap {
		fmt.Println(address) // Debug

		// To-Do: Add logic for re-redeems
		if address == cometBftAddress {
			var searchHeight int64 = 0
			for redeem := range redeems {
				if redeems[redeem].redeemedBlockHeight > searchHeight {
					searchHeight = redeems[redeem].redeemedBlockHeight
				}
			}

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
	var searchHeight int64 = 0
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
	LastTrackerHeight int
	tokenIdMap        map[string][]Validator_RedeemEvent
	addressMap        map[string][]Validator_RedeemEvent
	Startsig          chan string
}

// Create a new tracker object to track an evenblt.
func NewTracker(rpcSourceAddress string, rpcSearchLimit int, TrackedEvent Rpc_RedeemEvent) *Tracker {
	return &Tracker{
		RpcAddress:        rpcSourceAddress,
		rpcSearchLimit:    rpcSearchLimit,
		TrackedEvent:      TrackedEvent,
		ValidatorList:     []Validator_RedeemEvent{},
		tokenIdMap:        map[string][]Validator_RedeemEvent{},
		addressMap:        map[string][]Validator_RedeemEvent{},
		LastTrackerHeight: 0,
		Startsig:          make(chan string),
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
	startTime := time.Now()
	// Do a historical search of all redeem events from deployBlock to latestBlock
	list, err := nft_tracker.FindRedeems(nft_tracker.TrackedEvent.deployBlock, int(latestBlock)-confirmations)
	if err != nil {
		errChannel <- noLatestBlock
	}
	nft_tracker.Startsig <- "done"
	fmt.Println("Found", list, "redeem events in historical search, took time:", time.Since(startTime))
	latestCheckedBlock := int(latestBlock)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		fmt.Println("Checking for new blocks")
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
func (nft_tracker *Tracker) FindRedeems(fromBlock int, toBlock int) (int, error) {
	RedeemsFound := 0
	lastUpdate := 0
	if nft_tracker.rpcSearchLimit == 0 { // Unlimited RPC, no need to search incrementally.
		list, err := nft_tracker.FetchAppendRedeems(nft_tracker.TrackedEvent.deployBlock, toBlock)
		if err != nil {
			return 0, err
		}
		RedeemsFound = len(list)
	} else {
		for currentBlock := fromBlock; currentBlock <= toBlock; currentBlock += nft_tracker.rpcSearchLimit + 1 {
			// Search through all blocks incrementing by rpcSearchLimit.
			list, err := nft_tracker.FetchAppendRedeems(currentBlock, (currentBlock + nft_tracker.rpcSearchLimit))
			if err != nil {
				return 0, err
			}
			RedeemsFound += len(list)
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
			fmt.Println(ValidatorList[vpass].ToString())
			RedeemsFound = append(RedeemsFound, ValidatorList[vpass])
			nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])

			// Add to corresponding maps for tokenid and validator address
			nft_tracker.AddToTokenIdMap(ValidatorList[vpass])
			nft_tracker.AddToAddressMap(ValidatorList[vpass])
		}
		// Update nft_tracker.lastTrackerHeight
		nft_tracker.LastTrackerHeight = toBlock
	}
	return RedeemsFound, nil
}

func (nft_tracker *Tracker) AddToTokenIdMap(validatorRedeem Validator_RedeemEvent) {
	currentRedeem, exists := nft_tracker.tokenIdMap[validatorRedeem.tokenId]
	if !exists {
		currentRedeem = []Validator_RedeemEvent{}
		currentRedeem = append(currentRedeem, validatorRedeem)
		nft_tracker.tokenIdMap[validatorRedeem.tokenId] = currentRedeem
	} else {
		currentRedeem = append(currentRedeem, validatorRedeem)
		nft_tracker.tokenIdMap[validatorRedeem.tokenId] = currentRedeem
	}
}

func (nft_tracker *Tracker) AddToAddressMap(validatorRedeem Validator_RedeemEvent) {
	currentRedeem, exists := nft_tracker.addressMap[validatorRedeem.validatorAddress]
	if !exists {
		currentRedeem = []Validator_RedeemEvent{}
		currentRedeem = append(currentRedeem, validatorRedeem)
		nft_tracker.addressMap[validatorRedeem.validatorAddress] = currentRedeem
	} else {
		currentRedeem = append(currentRedeem, validatorRedeem)
		nft_tracker.addressMap[validatorRedeem.validatorAddress] = currentRedeem
	}
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
		"fromBlock": string(fmt.Sprintf("0x%x", fromBlock)), // fromBlock,
		"toBlock":   string(fmt.Sprintf("0x%x", toBlock)),   // toBlock,
		"address":   TrackedEvent.contractAddress,
		"topics": []string{
			TrackedEvent.EventSignature,
		},
	}

	// fmt.Printf("Searching for redeem event logs in blocks %d to %d\n", fromBlock, toBlock)
	response := make([]RedeemEventRpc, 1)
	// To-Do: Handle responses that are error messages, for example if the RPC is down.
	getLogsFailed := ethereum_client.Client().Call(&response, "eth_getLogs", RpcArguments)
	if getLogsFailed != nil {
		return nil, getLogsFailed
	}
	// response handling logic
	if len(response) > 0 {
		for val := range response {
			redeemEventsInRange = append(redeemEventsInRange, *NewValidatorRedeemEvent(response[val].Topics[1], response[val].Data, response[val].BlockNumber))
		}
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
