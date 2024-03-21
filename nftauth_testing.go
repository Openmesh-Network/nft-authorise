package nft_auth

import (
	"crypto/ecdsa"
	"testing"

	"github.com/nanmu42/etherscan-api"
)

const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
const sepolia etherscan.Network = "api-sepolia"

func TestDeriveAddress(t *testing.T) {
	// Test that the address is generated correctly
	supposedAddress := deriveAddressFromPubkey(ecdsa.PublicKey{})
	t.Log(supposedAddress)
	// Test against known valid addresses with their associated public key.
}

func TestGetValidNFTs(t *testing.T) {
	var mintedNFTs = getValidNFTsFromEtherscan(contractAddress, sepolia)
	for nft := range mintedNFTs {
		t.Log(mintedNFTs[nft].tokenId)
	}
}
