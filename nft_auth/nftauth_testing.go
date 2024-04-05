package nft_auth

import (
	"encoding/hex"
	"testing"

	ethereum "github.com/ethereum/go-ethereum/crypto"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"

//const sepolia etherscan.Network = "api-sepolia"

func TestNftTracker(t *testing.T) {
	// Compute Keccak256 hash of the event signature
	hash := ethereum.Keccak256([]byte("redeem(uint256)"))
	eventSignature := hex.EncodeToString(hash)

	// Concatenate "0x" with the event signature
	eventSignatureWithPrefix := "0x" + eventSignature

	t.Log("Event sig:", eventSignatureWithPrefix)
	trackerobj := NewTracker("https://rpc.ankr.com/eth_sepolia")
	trackerobj.Start(contractAddress)
}

/*
func TestGetValidNFTs(t *testing.T) {
	var mintedNFTs = getValidNFTsFromEtherscan(contractAddress, sepolia)
	for nft := range mintedNFTs {
		t.Log(mintedNFTs[nft].tokenId)
	}
}

// Independent test using etherscan on MockNFTs. Later, we should abstract an input for 'getValidNFTsFromFunc(fetchFunc func(), contractAddress string) []Validator_Pass'
func getValidNFTsFromEtherscan(contractAddress string, network etherscanapi.Network) []Validator_Pass {
	client := etherscanapi.New(network, "")
	client.ContractABI(contractAddress)

	//interest, err := client.TokenTotalSupply(contractAddress)

	walletAddress := "0x0000000000000000000000000000000000000000"
	validNFTs, err := client.ERC721Transfers(&contractAddress, &walletAddress, nil, nil, 0, 0, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(validNFTs)
	for validNFTs := range validNFTs {
		fmt.Println(validNFTs)
	}

	return nil
}

func TestMockNFT(t *testing.T) {

	// Test validating NFTs.
	// Get NFTs from Etherscan (sepolia contract)
	testValidator := ""
	validatorTestList := getValidNFTsFromEtherscan(contractAddress, sepolia)
	result := validateNFTMembership(testValidator, validatorTestList)
	t.Log("Result of NFT validation:", result)
}
*/
