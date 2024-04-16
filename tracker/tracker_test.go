package nft_auth

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
const deployBlock = 5517796 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.
const rpcSource = "https://rpc.ankr.com/eth_sepolia"

// To fix: Tests all run out of time. Try with own Ethereum RPC.
func TestNftTracker(t *testing.T) {
	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte("Redeemed(uint256,bytes32)"))
	eventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	eventSignatureWithPrefix := "0x" + eventSignature

	t.Log("Event sig:", eventSignatureWithPrefix)
	trackerobj := NewTracker(rpcSource)
	ctx := context.Background()
	trackerobj.StartTracking(ctx, contractAddress, deployBlock, 2*time.Minute, 20)
}
func TestRPCfetch(t *testing.T) {
	// Testing reveals that response is empty if there are no redeem events found.
	helper_FindVPassinRange(5618691, 5618693, t)

	// One Validator pass object found in this range. (Next check for a double redeem within 3 blocks.)
	helper_FindVPassinRange(5618693, 5618695, t)

	trackerobj := NewTracker(rpcSource)
	// Test longer range with historical fetch loop.
	trackerobj.helper_FindVPassHistorical(5618685, 5618705, t)

	// Test too long, use in main.go
	// helper_FindVPassHistorical(deployBlock, 5673530, t)
}

func TestCallbackfuncs(t *testing.T) {
	ctx := context.Background()
	trackerobj := NewTracker(rpcSource)
	trackerobj.StartTracking(ctx, "0x8D64aB58a17dA7d8788367549c513386f09a0A70", 5517796, 2*time.Minute, 20)

	redeemed := Verify_join_callback("61a83a39c806449ddc66feb6c86a1994456a8c8b", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

// Found Validator pass with token id:  0x0000000000000000000000000000000000000000000000000000000000000001 and validator address:  0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000

func TestSubscribeEthereumEvent(t *testing.T) {
	// Set up subscription
	quit := make(chan []byte)
	response := make(chan []byte)
	ctx := context.Background()

	// Set up Ethereum Client
	ethereum_client, err := ethclient.Dial(rpcSource)
	if err != nil {
		panic(err) // Should surface an error about the rpc source if possible
	}
	defer ethereum_client.Close()
	RpcArguments := map[string]interface{}{
		"address": contractAddress,
		"topics": []string{
			"0x4fc9c25b46f7854a495f8830e3d532a48cd64b4e4e3f6038557fe5669885bbe6", // Keccak256 hash, is the event signature of: Redeemed(uint256,bytes32)
		},
	}

	subscription, subscribeErr := ethereum_client.Client().Subscribe(ctx, "eth", <-response, RpcArguments)
	if subscribeErr != nil {
		panic(subscribeErr)
	}
	defer subscription.Unsubscribe()
	// Run loop from the StartTracking subscribe loop.
	for {
		select {
		case <-quit:
			t.Log("Received quit signal, stopping...")
			return // Exit the goroutine
		default:
			// Subscribe to event stream and update accordingly. SHOULDN'T RESUBSCRIBE EVERY SINGLE TIME! SUB ONCE, HANDLE RESPONSES IN A LOOP.
			if subscription.Err() != nil {
				panic(subscribeErr)
			}

			// Update the list of validator pass redeem events
			t.Log(response)
		}
	}
}

// Helpers
func helper_FindVPassinRange(toblock int, fromblock int, t *testing.T) {
	list, err := FetchRedeemEventsRPC(rpcSource, contractAddress, toblock, fromblock)
	if err != nil {
		panic(err)
	}
	if len(list) == 0 {
		t.Log("No NFTs found")
	}
	for vp := range list {
		t.Log("Found Validator pass with token id: ", list[vp].tokenId, "and validator address: ", list[vp].validatorAddress)
	}
	t.Log("Found", len(list), "NFTs")
}

func (nft_tracker *Tracker) helper_FindVPassHistorical(toblock int, fromblock int, t *testing.T) {
	list, err := nft_tracker.FetchHistoricalRedeems(rpcSource, contractAddress, toblock, fromblock)
	if err != nil {
		panic(err)
	}
	if len(list) == 0 {
		t.Log("No NFTs found")
	}
	for vp := range list {
		t.Log("Found Validator pass with token id: ", list[vp].tokenId, "and validator address: ", list[vp].validatorAddress)
	}
	t.Log("Found", len(list), "NFTs")
}
