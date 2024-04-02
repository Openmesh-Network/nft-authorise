package nft_auth

import (
	"fmt"
	"net/rpc"
	"time"
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

func Start(trackerSource Tracker, contractAddress string) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// The threat we would be tackling here is a malicious RPC source, which is difficult to defend against, but there should be some basic resilience.
				ValidatorList, err := fetchNFTsRPC(trackerSource, contractAddress)
				// We trust the previously acquired validator lists, if there are conflicts with new additions only trust the ones confirmed by the most RPC requests (over time)
				// Inconsistencies in RPC response should mean the list does not change for some time, wait until the new request-responses have built up to be greater than the old ones.
				if err != nil {
					panic(err)
				}
				// Implement error handling and behaviour for when RPC source fails.
				if len(ValidatorList) == 0 {
					// If chosen RPC source fails, try etherscan.
					fmt.Println("No NFTs found. Trying etherscan.")
					/*
						ValidatorList, err := getValidNFTsFromEtherscan(contractAddress, sepolia)
						if err != nil {
							panic(err)
						}
					*/
				}
				trackerSource.ValidatorList = ValidatorList

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

// Fetch a full list of Validator Passes from a smart contract address.
func fetchNFTsRPC(source Tracker, contractAddress string) ([]Validator_Pass, error) {
	// Get NFTs from RPC
	var validNFTs []Validator_Pass
	rpc.Dial("sepolia", source.RpcAddress)
	return validNFTs, nil
}

// May return "Nft not found" if not found in the list.
func getNFTfromValidatorPassList(tokenid string, validNFTs []Validator_Pass) string {
	for nft := range validNFTs {
		if validNFTs[nft].tokenId == tokenid {
			fmt.Println(validNFTs[nft].validatorAddress)
			return validNFTs[nft].validatorAddress
		}
	}
	fmt.Printf("Nft not found by tokenid: %s", tokenid)
	return "Nft not found"
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
