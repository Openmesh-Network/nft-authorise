package nft_auth

import (
	"encoding/hex"
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
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia")
	trackerobj.Start(contractAddress, deployBlock)
}

func TestRPCfetch(t *testing.T) {
	list, err := FetchValidatorPassesRPC("https://rpc.ankr.com/eth_sepolia", contractAddress, 5618693, 5618695)
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

func TestCallbackfuncs(t *testing.T) {
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia")
	trackerobj.Start("0x8D64aB58a17dA7d8788367549c513386f09a0A70", 5517796)

	redeemed := join_callback("61a83a39c806449ddc66feb6c86a1994456a8c8b", trackerobj)
	t.Log("Tracked a successful redeem for cometBFT address: ", redeemed)
}

// Found Validator pass with token id:  0x0000000000000000000000000000000000000000000000000000000000000001 and validator address:  0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000
