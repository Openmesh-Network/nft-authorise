package validatorpass_tracker

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
const deployBlock = 5517796 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.
// const rpcSource = "https://rpc.sepolia.org"
const rpcSource = "https://rpc.ankr.com/eth_sepolia"
const redeemed = "Redeemed(uint256,bytes32)"

var RedeemEvent = NewRedeemEvent(redeemed, contractAddress, deployBlock)

// To fix: Tests all run out of time. Try with own Ethereum RPC.
func TestNftTracker(t *testing.T) {
	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte(redeemed))
	EventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	EventSignatureWithPrefix := "0x" + EventSignature

	t.Log("Event sig:", EventSignatureWithPrefix)
	trackerobj := NewTracker(rpcSource, 4, RedeemEvent)
	ctx := context.Background()
	trackerobj.StartTracking(ctx, 2*time.Minute, 20)
}
func TestFindVPassRPC(t *testing.T) {
	// Testing reveals that response is empty if there are no redeem events found.
	FindVPassinRange(5618691, 5618693, t)

	// One Validator pass object found in this range. (Next check for a double redeem within 3 blocks.)
	FindVPassinRange(5618693, 5618750, t)

	trackerobj := NewTracker(rpcSource, 4, RedeemEvent)
	// Test longer range with historical fetch loop.
	trackerobj.FindVPassHistorical(5618685, 5618705, t)

	// Test live search using main.go
}

func TestCallbackfuncs(t *testing.T) {
	ctx := context.Background()
	trackerobj := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, 5618691))
	go func() {
		trackerobj.StartTracking(ctx, 2*time.Minute, 20)
	}()
	time.Sleep(3 * time.Second)
	redeemed := VerifyAddress("0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

func TestFetchRPC(t *testing.T) {
	nft_tracker := NewTracker(rpcSource, 4, RedeemEvent)
	RedeemsFound := []Validator_RedeemEvent{}
	// Ripped straight from FetchAppendRPC, lazy test
	ValidatorList, err := nft_tracker.FetchAppendRedeems(5618693, 5618695) // Hardcode test values.
	if err != nil {
		t.Log(err)
	}
	if len(ValidatorList) > 0 {
		for vpass := range ValidatorList {
			RedeemsFound = append(RedeemsFound, ValidatorList[vpass])
			nft_tracker.ValidatorList = append(nft_tracker.ValidatorList, ValidatorList[vpass])
		}
	}
	for r := range RedeemsFound {
		//fmt.Println(RedeemsFound[r].ToString())
		fmt.Println(RedeemsFound[r].validatorAddress)
	}
}

func TestMaximumBlockRange(t *testing.T) {
	ethereum_client, err := ethclient.Dial(rpcSource)
	if err != nil {
		panic(err) // Crash if no access to RPC.
	}
	defer ethereum_client.Close()
	list, err := FetchRedeemEventsRPC(rpcSource, RedeemEvent, 5618691, 5618695)
	if err != nil {
		t.Log(err)
	}
	if len(list) > 0 {
		for i := range list {
			print(list[i].ToString())
		}
	}
}

func TestHex(t *testing.T) {
	fromBlock := 5729623
	t.Log(string(fmt.Sprintf("0x%x", fromBlock)))
}

func TestUnlimitedBlockRange(t *testing.T) {
	trackerobj := NewTracker(rpcSource, 0, NewRedeemEvent(redeemed, contractAddress, deployBlock))
	// Test longer range with historical fetch loop.
	trackerobj.FindVPassHistorical(5618693, 5729623, t)

	trackerobj.StartTracking(context.Background(), 2*time.Second, 20)
}

func TestAddToMap(t *testing.T) {
	trackerobj := NewTracker(rpcSource, 4, NewRedeemEvent(redeemed, contractAddress, 5618691))
	trackerobj.AddToAddressMap(*NewValidatorRedeemEvent("0x0000000000000000000000000000000000000000000000000000000000000001", "0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000", "5618691"))
	t.Log(trackerobj.addressMap["0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000"][0].ToString())
}

// /////////////////// Helper functions /////////////////////
func FindVPassinRange(toblock int, fromblock int, t *testing.T) {
	list, err := FetchRedeemEventsRPC(rpcSource, NewRedeemEvent(redeemed, contractAddress, deployBlock), toblock, fromblock)
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
