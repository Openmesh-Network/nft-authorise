package validatorpass_tracker

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/crypto"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
const deployBlock = 5517796 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.
const rpcSource = "https://rpc.ankr.com/eth_sepolia"
const redeemed = "Redeemed(uint256,bytes32)"

// To fix: Tests all run out of time. Try with own Ethereum RPC.
func TestNftTracker(t *testing.T) {
	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte(redeemed))
	EventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	EventSignatureWithPrefix := "0x" + EventSignature

	t.Log("Event sig:", EventSignatureWithPrefix)
	trackerobj := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, deployBlock))
	ctx := context.Background()
	trackerobj.StartTracking(ctx, 2*time.Minute, 20)
}
func TestRPCfetch(t *testing.T) {
	// Testing reveals that response is empty if there are no redeem events found.
	FindVPassinRange(5618691, 5618693, t)

	// One Validator pass object found in this range. (Next check for a double redeem within 3 blocks.)
	FindVPassinRange(5618693, 5618695, t)

	trackerobj := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, deployBlock))
	// Test longer range with historical fetch loop.
	trackerobj.FindVPassHistorical(5618685, 5618705, t)

	// Test live search using main.go
}

func TestCallbackfuncs(t *testing.T) {
	ctx := context.Background()
	trackerobj := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, deployBlock))
	trackerobj.StartTracking(ctx, 2*time.Minute, 20)

	redeemed := VerifyMembershipOfAddress("61a83a39c806449ddc66feb6c86a1994456a8c8b", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

func TestFetchRPC(t *testing.T) {
	nft_tracker := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, deployBlock))
	RedeemsFound := []Validator_RedeemEvent{}
	// Ripped straight from FetchAppendRPC, lazy test
	ValidatorList, err := FetchRedeemEventsRPC(nft_tracker.RpcAddress, nft_tracker.TrackedEvent, 5618693, 5618695) // Hardcode test values.
	if err != nil {
		t.Log(err)
	}
	if len(ValidatorList) > 0 {
		for vpass := range ValidatorList {
			//fmt.Println(ValidatorList[vpass].validatorAddress)
			RedeemsFound = append(RedeemsFound, ValidatorList[vpass])
			nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])
		}
	}
	for r := range RedeemsFound {
		fmt.Println(RedeemsFound[r].ToString())
	}
}

// /////////////////// Helper functions /////////////////////
func FindVPassinRange(toblock int, fromblock int, t *testing.T) {
	redeemed := NewRedeemEvent(redeemed, contractAddress, deployBlock)
	list, err := FetchRedeemEventsRPC(rpcSource, redeemed, toblock, fromblock)
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
	list, err := nft_tracker.FetchAppendRedeems(toblock, fromblock)
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
