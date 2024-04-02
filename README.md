# nft-authorise
Authentication library to confirm public-key metadata associated with an NFT via a signing challenge.

The aim with this library is to be a modular, portable SDK for soul-bound NFT authentication that can facilitate a handshake protocol whereby the receiver knows for certain that the send is the wallet owner of an NFT from the specified contract. This library is intended for DAO's which use NFT as an authentication mechanism.

The verify function works with the following parameters: Contract Address (NFT Token Mint), Claimed Wallet Address (containing NFT). And possibly a public key or CometBFT meta-data attached the NFT.

## NFT Handshake in Openmesh Network - Example Implementation
User needs to: 

Generate / read their CometBFT address from Xnode Studio.

Redeem NFT with CometBFT address, costs ETH gas (provide UI in studio).

All Validators are regularly asking ETH RPC for transactions associated with the VPass contract.

Should keep a list of Wallets holding NFTs.

Keeps a list of TokenIDs and updates with CBFT address after they redeem.

Event sourcing:

Subscribe to the event.

getLogs to get all.

Ranking of Ethereum sources 

[Openmesh Dedicated, Etherscan, Ankr]

Removing peers

Voting power is set to 0 if a new redeem event for the same tokenId.

When reading an event, if tokenId is already in the list then remove the original CBFT address.

https://sepolia.etherscan.io/address/0x8d64ab58a17da7d8788367549c513386f09a0a70#writeContract
