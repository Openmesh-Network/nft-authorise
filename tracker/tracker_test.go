package nft_auth

import (
	"encoding/hex"
	"reflect"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/crypto"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
const deployBlock = 5517796 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.

func TestNftTracker(t *testing.T) {
	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte("Redeemed(uint256,bytes32)"))
	eventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	eventSignatureWithPrefix := "0x" + eventSignature

	t.Log("Event sig:", eventSignatureWithPrefix)
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia", false)
	trackerobj.StartTracking(contractAddress, deployBlock, nil)
}

func helper_FindVPassinRange(toblock int, fromblock int, t *testing.T) {
	list, err := FetchRedeemEventsRPC("https://rpc.ankr.com/eth_sepolia", contractAddress, toblock, fromblock)
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

func TestRPCfetch(t *testing.T) {
	// Testing reveals that response is empty if there are no redeem events found.
	helper_FindVPassinRange(5618691, 5618693, t)

	// One Validator pass object found in this range. (Next check for a double redeem within 3 blocks.)
	helper_FindVPassinRange(5618693, 5618695, t)
}

func TestCallbackfuncs(t *testing.T) {
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia", false)
	trackerobj.StartTracking("0x8D64aB58a17dA7d8788367549c513386f09a0A70", 5517796, nil)

	redeemed := join_callback("61a83a39c806449ddc66feb6c86a1994456a8c8b", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

// Found Validator pass with token id:  0x0000000000000000000000000000000000000000000000000000000000000001 and validator address:  0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000

func TestSaveLoad(t *testing.T) {
	path := "../test_gob"
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia", true)
	trackerobj.ValidatorList = []Validator_RedeemEvent{{"", "", 0}, {"", "", 0}}

	trackerobj.SaveToFile(path)
	newTrackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia", true)
	newTrackerobj.LoadFromFile(path)
}

func TestTypeComparisonGo(t *testing.T) {
	path := "../string"
	if reflect.TypeOf(path) == reflect.TypeFor[string]() {
		t.Log(path, "(passed)")
	} else {
		t.Log("Didn't work")
	}
	badpath := 5555
	if reflect.TypeOf(badpath) == reflect.TypeFor[string]() {
		t.Log(badpath)
	} else {
		t.Log("int not a string (passed)")
	}
}
