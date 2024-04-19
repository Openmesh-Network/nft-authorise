package nft_auth

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/crypto"
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
	FindVPassinRange(5618691, 5618693, t)

	// One Validator pass object found in this range. (Next check for a double redeem within 3 blocks.)
	FindVPassinRange(5618693, 5618695, t)

	trackerobj := NewTracker(rpcSource)
	// Test longer range with historical fetch loop.
	trackerobj.FindVPassHistorical(5618685, 5618705, t)

	// Test live search (latest block - 50) for the latest redeem events, should be tested right after an on-chain redeem
}

func TestCallbackfuncs(t *testing.T) {
	ctx := context.Background()
	trackerobj := NewTracker(rpcSource)
	trackerobj.StartTracking(ctx, "0x8D64aB58a17dA7d8788367549c513386f09a0A70", 5517796, 2*time.Minute, 20)

	redeemed := VerifyMembershipOfAddress("61a83a39c806449ddc66feb6c86a1994456a8c8b", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

// Helpers
func FindVPassinRange(toblock int, fromblock int, t *testing.T) {
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

func (nft_tracker *Tracker) FindVPassHistorical(toblock int, fromblock int, t *testing.T) {
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
