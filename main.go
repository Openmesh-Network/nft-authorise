package main

import (
	"context"
	"fmt"
	"time"

	vpauth "github.com/Openmesh-Network/nft-authorise/tracker"
)

func main() {
	// RPC constants
	const contractAddress string = "0x8D64aB58a17dA7d8788367549c513386f09a0A70"
	const rpcSource = "https://rpc.ankr.com/eth_sepolia"
	const deployBlock = 5618691 // 0x55bc06 This is the block at which the validator pass contact was deployed on-chain.

	// Concatenate "0x" with the event signature

	trackerobj := vpauth.NewTracker(rpcSource, 3000, vpauth.NewRedeemEvent("Redeemed(uint256,bytes32)", contractAddress, deployBlock))
	fmt.Printf("Tracking event with signature: %s\n", trackerobj.TrackedEvent.EventSignature)
	ctx := context.Background()

	trackerobj.StartTracking(ctx, 2*time.Minute, 20)

}

/*
	// Run tracking inside a go routine so that you can run checks on it every 2 minutes.
	go func() {
		trackerobj.StartTracking(ctx, contractAddress, deployBlock, 2*time.Minute, 20)
		defer trackerobj.Stop()
	}()

	// Ask the tracker about what's happening on ethereum using cometbft callbacks every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-ticker.C:
			askAboutRedeems(trackerobj)
		}
	}
}
// Testing function
func askAboutRedeems(tracker *vpauth.Tracker) {
	// Ask the tracker about what's happening on ethereum using cometbft callbacks every 2 minutes
	fmt.Println("Found first redeem: ", vpauth.Verify_join_callback("0x61a83a39c806449ddc66feb6c86a1994456a8c8b000000000000000000000000", tracker))
	fmt.Println("Found second redeem: ", vpauth.Verify_join_callback("0x2757295701725127590000000000000000000000000000000000000000000000", tracker))
	fmt.Println("Found third redeem: ", vpauth.Verify_join_callback("0x2175091590317500000000000000000000000000000000000000000000000000", tracker))

	// vpauth.Verify_ValidatorRedeemEvent_callback()
	fmt.Println("Length of validator list (including re-redeems):", len(tracker.ValidatorList))
*/
