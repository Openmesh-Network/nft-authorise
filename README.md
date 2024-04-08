# nft-authorise
Authentication library to confirm public-key metadata associated with an NFT via a signing challenge.

The aim with this library is to be a modular, portable SDK for soul-bound NFT authentication that can facilitate a handshake protocol whereby the receiver knows for certain that the send is the wallet owner of an NFT from the specified contract. This library is intended for DAO's which use NFT as an authentication mechanism.

The verify function works with the following parameters: Contract Address (NFT Token Mint), Claimed Wallet Address (containing NFT). And possibly a public key or CometBFT meta-data attached the NFT.

## NFT Handshake in Openmesh Network - Example Implementation
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
