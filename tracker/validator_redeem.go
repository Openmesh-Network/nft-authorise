package validatorpass_tracker

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// VALIDATOR REDEEM EVENTS

type Validator_RedeemEvent struct { // Optimisation: store strings as bytes[32] if memory is an issue
	tokenId             string // NFT token ID
	validatorAddress    string // CometBFT validator address
	redeemedBlockHeight int64  // Block height at which the validator pass was redeemed
}

// Records validator pass redeem events including the redeemed validator address and the the block height at which it was redeemed.
// All parameters are in form of "0x0000000000000000000000000000000000000000000000000000000000000001" as returned by RPC.
// redeemedBlockHeight assumes a string response from JSON RPC in form "0xabc".
func NewValidatorRedeemEvent(tokenId string, validatorAddress string, redeemedBlockHeight string) *Validator_RedeemEvent {
	heightNumerical, err := strconv.ParseInt(redeemedBlockHeight, 0, 0)
	if err != nil {
		fmt.Println("Unable to convert blockNumber response to integer: ", redeemedBlockHeight, err)
		heightNumerical = 0
	}
	return &Validator_RedeemEvent{
		tokenId:             tokenId,
		validatorAddress:    validatorAddress,
		redeemedBlockHeight: heightNumerical,
	}
}

func (vRedeem *Validator_RedeemEvent) ToString() string {
	return fmt.Sprintf("TokenId: %s, Validator Address: %s, Redeemed@Height: %d", vRedeem.tokenId, vRedeem.validatorAddress, vRedeem.redeemedBlockHeight)
}

// RPC Redeem events that we are interested in and what contract they are associated to.
type Rpc_RedeemEvent struct {
	EventSignature  string // Redeemed(uint256,bytes32)
	contractAddress string
	deployBlock     int
}

// Function for initialising the ethereum events you are interested in tracking, requires event, contract address and deploy block.
// Pass the event in format: function(datatype1,datatype2)
// eg. "Redeemed(uint256,bytes32)"
// This function mainly serves to create the input required for rpc interaction or to create a new tracker object.
func NewRedeemEvent(event string, contractAddress string, deployBlock int) Rpc_RedeemEvent {
	return Rpc_RedeemEvent{
		EventSignature:  GetEventSignature(event),
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
// If passed a string starting with 0x, it will assume this is the event signature and simply reflect it.
func GetEventSignature(eventString string) string {
	if strings.HasPrefix(eventString, "0x") {
		return eventString
	} else {
		// Compute Keccak256 hash of the event signature
		hash := crypto.Keccak256([]byte(eventString))
		EventSignature := hex.EncodeToString(hash)
		// Prefix needed for interpretation in RPC
		EventSignatureWithPrefix := "0x" + EventSignature
		return EventSignatureWithPrefix
	}
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
