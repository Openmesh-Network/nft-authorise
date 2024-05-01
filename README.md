# nft-authorise
Authentication library to find and record wallet address metadata associated with an NFT, to authorise validator's addresses when they request to join a permissioned network.

The aim with this library is to provide a deterministic callback function for peers joining a CometBFT-based network. It does this by 'tracking' a contract address for an event signature, which in the test-case is a *Redeemed* event on the initial *Mock Validator Pass* contract.

The verify function is a callback that checks the tracker's list of validator redeem events for a tokenId and confirms that the input Validator Address matches the latest redeem event for that token.

## Optimisations
Currently, the tracker keeps all redeem events an array in memory. This serves fine with our scope 1-10K validators, however if the network scales further it could be ideal to use a keystore to save resources. 

## Security Improvements
The primary security concern with this authentication library is that you trust the Ethereum RPC source implicitly, so it is recommended to run an ethereum node (lite or full is fine) on the local machine to use for these requests. 


To-Do:
* Ensure compatibility with CometBFT playback - historical redeems that are no longer valid should still be replayable.
* Unlimited RPC request configuration option
    * Develop and test using a Lite and Full node, in case the user has a locally hosted RPC source.
* Removing validators when there is a new redeem event to a different address.

## Use-case: Validator Passes in the Openmesh Network
The user journey for authenticating a validator node is simplified to the following steps:

1. Initialise a validator node and record it's address.

2. Redeem a Validator Pass with the node's CometBFT address.

In order to replay or verify blocks in Openmesh Core, nodes are regularly tracking RPC for redeem events associated with the Validator Pass contract. 

### Event sourcing
The RPC source is configurable when creating the tracker object by passing the URL as a parameter. Ethereum or Polygon RPC is expected, see main.go for an example program. However, any implementation is intended to be through importing the package rather than running this as a program.

The eth_getLogs rpc call is made repeatedly to search through blocks of any range with the assumption (based on Ankr public limit) that the RPC will only allow a search of 4 blocks at a time. 

### Removing peers

Voting power is set to 0 if a new redeem event for the same tokenId.

When reading an event, if tokenId is already in the list then remove the original CBFT address.

https://sepolia.etherscan.io/address/0x8d64ab58a17da7d8788367549c513386f09a0a70#writeContract
