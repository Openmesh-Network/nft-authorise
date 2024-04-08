package main

import (
	"encoding/hex"
	"fmt"

	vpauth "github.com/Openmesh-Network/nft-authorise/nft_auth"
	ethereum "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
	const rpcSource = "https://rpc.ankr.com/eth_sepolia"
	const deployBlock = 5517796 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.

	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte("cheapRedeem(uint256,bytes32)"))
	eventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	eventSignatureWithPrefix := "0x" + eventSignature

	fmt.Printf("Event sig: %s\n", eventSignatureWithPrefix)
	trackerobj := vpauth.NewTracker(rpcSource)
	fmt.Println("Tracker object created.")
	trackerobj.Start(contractAddress, deployBlock)
	defer trackerobj.Stop()
	fmt.Println("Tracker started, waiting for go routine.")
	ValidatorList, err := vpauth.FetchValidatorPassesRPC(rpcSource, contractAddress, string(deployBlock), string(deployBlock+5))
	if err != nil {
		panic(err)
	}
	if len(ValidatorList) != 0 {
		fmt.Println(ValidatorList[0])
	}
}
