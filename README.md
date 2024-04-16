# nft-authorise
Authentication library to confirm public-key metadata associated with an NFT via a signing challenge. For the NFT mock contract which this library is based on,

The aim with this library is to be a modular, portable SDK for NFT authentication that can facilitate a handshake protocol whereby the receiver knows for certain that the send is the wallet owner of an NFT from the specified contract. This library is built to keep track of NFTs as an authentication mechanism, it is developed to provide a deterministic callback function for peers joining a CometBFT-based network.

The verify function works with the following parameters: Contract Address (NFT Token Mint), Claimed Wallet Address (containing NFT). And possibly a public key or CometBFT meta-data attached the NFT.

The primary security concern with this authentication library is that you trust the Ethereum RPC source implicitly, so it is recommended to run an ethereum node (lite or full is fine) on the local machine to use for these requests. 

To-Do:
* Change string values in the Redeem Event struct to hex / binary to save space in memory and storage.
* Ensure compatibility with CometBFT playback (historical redeems that are no longer valid should still be replayable).
* Error handling and program resilience (for example, check what happens if we hit the rate limit).
* Unlimited RPC requests, need to run an ethereum full node for developing this.

## Validator Pass in the Openmesh Network
User needs to: 

1. Generate / read their CometBFT address from Xnode Studio.

2. Redeem NFT with CometBFT address, costs ETH gas (provide UI in studio).

All Validators are regularly tracking ETH RPC for redeem events associated with the Validator Pass contract. This tracker keeps a list of valid Comet BFT address' per tokenId / redeem event.

#### Event sourcing
The RPC source is configurable when creating the tracker object by passing the URL as a parameter. It must be either Ethereum or Polygon RPC to be compatible, see an example in Openmesh Core (at the time of writing this in feat/nft-tracker branch).

Subscribe to the redeem event. An RPC subscription is maintained by all nodes for the purpose of verifying a transaction / determining voting power in the permissioned network.

getLogs is required by a block proposer or called over a long period to check for any missed redeem events.


#### Removing peers (Future)

Voting power is set to 0 if a new redeem event for the same tokenId.

When reading an event, if tokenId is already in the list then remove the original CBFT address.

https://sepolia.etherscan.io/address/0x8d64ab58a17da7d8788367549c513386f09a0a70#writeContract
