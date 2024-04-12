package nft_auth

// VALIDATOR REDEEM EVENTS

type Validator_RedeemEvent struct {
	tokenId             string // NFT token ID
	validatorAddress    string // CometBFT validator address eg. cometvaloper1abc123def456ghi789jkl123mno456pqr789stu
	redeemedBlockHeight int    // Block height at which the validator pass was redeemed
}

func NewValidatorRedeemEvent(tokenId string, validatorAddress string) *Validator_RedeemEvent {
	return &Validator_RedeemEvent{
		tokenId:          tokenId,
		validatorAddress: validatorAddress,
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
