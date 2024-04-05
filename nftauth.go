package nft_auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Validator_Pass struct {
	tokenId          string // NFT token ID
	validatorAddress string // CometBFT validator address eg. cometvaloper1abc123def456ghi789jkl123mno456pqr789stu
}

type Tracker struct {
	// Abstracts the source of Ethereum RPC to be used for NFT tracking. (Support for external or internal nodes)
	RpcAddress    string
	ValidatorList []Validator_Pass
}

func NewTracker(rpcSourceAddress string) *Tracker {
	fmt.Println(rpcSourceAddress)
	return &Tracker{
		RpcAddress:    rpcSourceAddress,
		ValidatorList: []Validator_Pass{},
	}
}

// Start tracking redeem events from a Validator Pass smart contract address, you should be able to deterministically call validateNFTMembership() for peer validation in a CometBFT callback.
func (nft_tracker *Tracker) Start(contractAddress string) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// Vulnerability: does not consider that the RPC source may be malicious, this source must be trusted!
				// To-Do: Use Openmesh verified Ethereum data as an RPC source.
				ValidatorList, err := fetchNFTsRPC(nft_tracker.RpcAddress, contractAddress)
				if err != nil {
					panic(err)
				}

				nft_tracker.ValidatorList = ValidatorList

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// Fetch a full list of Validator Passes from a smart contract address.
func fetchNFTsRPC(rpcSource string, contractAddress string) ([]Validator_Pass, error) {
	var validNFTs []Validator_Pass

	ethereum_client, err := ethclient.Dial(rpcSource)
	if err != nil {
		return nil, err
	}

	/*
		"params":
		[
			{
			"fromBlock": "0xCD0E64",
			"toBlock":"latest",
			"address":"0x514910771af9ca656af840dff83e8264ecf986ca",
			"topics":[
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
			]
			}
		],
	*/
	RpcArguments := []interface{}{
		map[string]interface{}{
			"fromBlock": "earliest",
			"toBlock":   "latest",
			"address":   "0x8D64aB58a17dA7d8788367549c513386f09a0A70",
			/*&"topics": []string{
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				// Unsure how to determine a topic, may need a second step to fetch topics if leaving them out doesn't fetch * all.
			},*/
		},
	}

	go func() {
		defer ethereum_client.Close()

		ctxToPreventHanging, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()
		fmt.Println("Waiting for ")
		resp := ethereum_client.Client().Call(ctxToPreventHanging, "eth_getLogs", RpcArguments)
		if err != nil {
			panic(err)
		}
		fmt.Println(resp)
	}()
	return validNFTs, nil
}

// May return "Nft not found" if not found in the list, looks for the latest CometBFT address associated to the NFT
func (nft_tracker *Tracker) checkCometBFTAddress(tokenid string) string {

	for nft := range nft_tracker.ValidatorList {
		if nft_tracker.ValidatorList[nft].tokenId == tokenid {
			fmt.Println(nft_tracker.ValidatorList[nft].validatorAddress)
			return "true"
		}
	}
	fmt.Printf("Nft not found by tokenid: %s", tokenid)
	return "false"
}

// Validate that there is an NFT with that public key in the list of valid NFTs. Very slow and inefficient method.
func validateNFTMembership(cometbftAddress string, validNFTs []Validator_Pass) bool {
	for nft := range validNFTs {
		if validNFTs[nft].validatorAddress == cometbftAddress {
			return true
		}
	}
	return false
}
