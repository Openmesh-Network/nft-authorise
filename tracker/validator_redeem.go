package validatorpass_tracker

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

// VALIDATOR REDEEM EVENTS

type Validator_RedeemEvent struct {
	tokenId             string // NFT token ID
	validatorAddress    string // CometBFT validator address
	redeemedBlockHeight int    // Block height at which the validator pass was redeemed
}

func NewValidatorRedeemEvent(tokenId string, validatorAddress string) *Validator_RedeemEvent {
	return &Validator_RedeemEvent{
		tokenId:          tokenId,
		validatorAddress: validatorAddress,
	}
}

// RPC Redeem events that we are interested in and what contract they are associated to.
type Rpc_RedeemEvent struct {
	eventSignature  string
	contractAddress string
	deployBlock     int
}

// Function for initialising the ethereum events you are interested in tracking, requires event, contract address and deploy block.
// Pass the event in format: function(datatype1,datatype2)
// eg. "Redeemed(uint256,bytes32)"
// This function mainly serves to create the input required for rpc interaction or to create a new tracker object.
func NewRedeemEvent(event string, contractAddress string, deployBlock int) Rpc_RedeemEvent {
	return Rpc_RedeemEvent{
		eventSignature:  GetEventSignature(event),
		contractAddress: contractAddress,
		deployBlock:     deployBlock,
	}
}

type RedeemEventRpc struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

// Get the event signature of an event, eg. "Redeemed(uint256,bytes32)"
func GetEventSignature(eventString string) string {
	// Compute Keccak256 hash of the event signature
	hash := crypto.Keccak256([]byte(eventString))
	eventSignature := hex.EncodeToString(hash)
	return eventSignature
}

// REDUNDANT FUNCTION
// Not used, no purpose unless you need to implement re-redeem functionality for you validator pass.
func UpdateRedeemEvent(newRedeem Validator_RedeemEvent, redeemList []Validator_RedeemEvent) (updatedList []Validator_RedeemEvent) {
	for redeem := range redeemList { // For each redeem if for the same tokenId
		if redeemList[redeem].tokenId == newRedeem.tokenId {
			// Double check that the newRedeem has a higher recorded block height.
			if newRedeem.redeemedBlockHeight > redeemList[redeem].redeemedBlockHeight {
				updatedList = append(updatedList, newRedeem)
			}
		} else { // Add events for new tokenId and keep events relating to unique tokenIds (eg. not the same as newRedeem's)
			updatedList = append(updatedList, redeemList[redeem])
		}
	}
	return updatedList
}
