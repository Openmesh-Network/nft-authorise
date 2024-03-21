package nft_auth

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	ethcrypt "github.com/ethereum/go-ethereum/crypto"
	etherscanapi "github.com/nanmu42/etherscan-api"
)

type Validator_Pass struct {
	tokenId          string // NFT token ID
	validatorAddress string // CometBFT validator address eg. cometvaloper1abc123def456ghi789jkl123mno456pqr789stu
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

// Hardcoded for etherscan as source of ethereum data.
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

// Derive the cometBFT address based on the public key (by hashing it)
func deriveAddressFromPubkey(claimedPublicKey ecdsa.PublicKey) string {
	// Convert the public key to uncompressed format
	uncompressedPubkey := ethcrypt.FromECDSAPub(&claimedPublicKey)

	// Take the Keccak-256 hash of the uncompressed public key
	hash := ethcrypt.Keccak256(uncompressedPubkey[1:])

	// Take the last 20 bytes of the hash as the wallet address
	address := hash[12:]

	// Convert the address to a hexadecimal string and add the "0x" prefix
	derivedWalletAddress := "0x" + hex.EncodeToString(address)

	return derivedWalletAddress
}

// Validate that there is an NFT with that public key in the list of valid NFTs. Very slow and inefficient method.
func validateNFTMembership(walletAddress string, claimedPublicKey ecdsa.PublicKey, validNFTs []Validator_Pass) bool {
	for nft := range validNFTs {
		if validNFTs[nft].validatorAddress == walletAddress {
			return true
		}
	}
	return false
}
