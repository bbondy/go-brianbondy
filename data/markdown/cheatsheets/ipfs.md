# IPFS

Labeling content by its content instead of where it is located.
Centralized web relies on trust, the decentralized web is trustless.
It is self verifiable by identifying content by its hash.
We can and do host each other's data


## Command line

### Install IPFS:

https://docs.ipfs.io/guides/guides/install/

### Start the daemon:

```
ipfs daemon
```

### Adding a file:

```
ipfs add ~/Downloads/image.jpg
```

### Adding a folder:

This adds not only the files in the folders recursively, but mirrors all directory structure as well.

```
ipfs add -r folder-name
```

> added QmSCq95Tce98WYWVqLG5L5Ev6XkaxxTZzTCFSMhRLVUUpa image.jpg

### Obtaining a file:

```
ipfs get QmSCq95Tce98WYWVqLG5L5Ev6XkaxxTZzTCFSMhRLVUUpa
```

### Obtaining a directory:

The same as for a file, but it will mirror the whole directory and files.

You can navigate a mirrored folder by using its CID through an IPFS gateway:

https://gateway.pinata.cloud/ipfs/<CID>

To access IPNS through the gateway:
- https://gateway.pinata.cloud/ipns/<peer-id>
- https://gateway.pinata.cloud/ipns/<dns-link-txt-record-domain>
- https://gateway.pinata.cloud/ipns/<dns-link-txt-record-domain>/images/someImage.jpg

If you have a daemon running locally you can also navigate from your local server:
- http://localhost:8080/ipfs/<CID>
- http://localhost:8080/ipns/<peer-id>
- http://localhost:8080/ipns/<dns-link-txt-record-domain>
- http://localhost:8080/ipns/<dns-link-txt-record-domain>/images/someImage.jpg


If you update ANY content, the CID changes though.


### Obtaining an object's metadata:

```
ipfs object get QmcQrDGFUyHdU5seZSNz6XH9dF6sM4A6JXpJAuvKzKZUUP
```

```
{"Links":[{"Name":"1","Hash":"QmSfsiv4Dd7vdfV6Rupd5dLFnuGZm24kCK8uHMa9DZ3xwQ","Size":87585},{"Name":"2","Hash":"QmSfsiv4Dd7vdfV6Rupd5dLFnuGZm24kCK8uHMa9DZ3xwQ","Size":87585}],"Data":"\u0008\u0001"}
```

> Saving file(s) to QmSCq95Tce98WYWVqLG5L5Ev6XkaxxTZzTCFSMhRLVUUpa
>  85.46 KiB / 85.46 KiB [========================================================] 100.00% 0s

### Mirror a website:

```
wget --mirror --convert-links --adjust-extension --page-requisites --no-parent <a href="https://vaibhavsaini.com" target="_blank">https://vaibhavsaini.com</a>
```


## Definitions

### Hashing in IPFS

IPFS uses SHA256 base58 encoded
IPFS uses a unique content identifier (CID) to address content.
IPFS has deduplication

### Content identifier (CID)

- A difference in content makes a different CID
- Using the same content, the same CID will be produced, regardless of its physical location
- There is a CID specification here: https://github.com/multiformats/cid
- There is a CID v0 and v1
    - v0: Original CID, base 58-encoded multihashes as the content identifiers. 46 characters and always starts with Qm.
    - v1: multibase prefix to identify the remainder of the CID, CID version identifier, multicodec identifier which indicates the format of the content.
- You can get info about how a CID is broken down by using:  https://cid.ipfs.io/#QmSCq95Tce98WYWVqLG5L5Ev6XkaxxTZzTCFSMhRLVUUpa or the CID explorer: https://explore.ipld.io/#/explore/QmSCq95Tce98WYWVqLG5L5Ev6XkaxxTZzTCFSMhRLVUUpa


Large files are chunked into 256 kB chunks, hashed, and organized into an IPLD

### Multihash

- Define hashes that are self defining
- Hash Type | Hash Value
 

### InterPlanetary Linked Data (IPLD)
- https://ipld.io/
- IPLD is the data model of the content-addressable webUses directed acyclic graphs (Merkle DAG) for managing the chunks and linking it to its CID
- Allow us to do for data what URLs and links did for HTML web pages (Bitcoin transaction, Ethereum transactions, Git commit, etc...).
- Links can be traversed across protocols, allowing you to explore data regardless of the underlying protocol
- IPLD is a single namespace for all hash-inspired protocols
- There are libraries for working with IPLD: https://github.com/ipld
    - There are other ipld libraries as well:
        - ipld-git
        - ipld-bitcoin
        - ipld-zcash
- Consists of 2 parts:
    - Data
    - Links - Array of IPLD links
- IPLD links consists of 3 parts
    - Name of the LinkHash
    - Hash of the linked IPFS object size
    - Cumulative size of the linked IPFS object

### Merkle trees:

- A hash tree, where every node is a hash
- Nodes point to other nodes by their hash
- A parent node's hash is composed of its child nodes
- The root node uniquely identifies all subcontent

### Multiformats:

- Allows you to future proof, the protocols have self describing formats to do this.
- Multiformats is composed of: 
    - multihash: Self describing hashes. Follows Type-Length-Value format (TLV)
        - Type is from here: https://github.com/multiformats/multicodec/blob/master/table.csv
        - The length is an unsigned integer describing the number of bytes to follow
        - The digest raw bytes 
        - Example: 122041dd....    0x12 is for sha2-256, 0x20 is 32
        - Several multihash libraries for various languages exist: http://multiformats.io/multihash/#implementations
    - multiaddr: Self describing addresses
        - Examples of things that are not multiaddr:
			- `[::1]:3217` -  IPv6. is this TCP? or UDP? or something else?
			- `127.0.0.1:9090` - IPv4. is this TCP? or UDP? or something else?
			- HTTP defaults to TCP
		- Examples of things that are multiaddr:
			- `/ip4/127.0.0.1/udp/9090/quic`
			- `/ip6/::1/tcp/3217`
			- `/ip4/127.0.0.1/tcp/80/http/baz.jpg`
			- `/dns4/foo.com/tcp/80/http/bar/baz.jpg`
			- `/dns6/foo.com/tcp/443/https`
		- It has both a human readable and binary packed version
		- Various language implementations here: http://multiformats.io/multiaddr/#implementations
    - multibase: Self describing base encodings
        - Disambiguate base32, base64, base58, etc
		- The format is: <base-encoding-character><base-encoded-data>
		- Encoding character is: https://github.com/multiformats/multibase/blob/master/multibase.csv
		- Multibase implementations: https://github.com/multiformats/multibase#implementations
    - multicodec: Self describing serialization
		- <multicodec><encoded-data>
		- multicodec protocol table: https://github.com/multiformats/multicodec/blob/master/table.csv
    - multistream:  Deprecated for multistream-select and multicodec
        - <varint-len>/<codec>\n<encoded-data>
    - multistream-select: Friendly protocol multiplexing
    - multigram: Self describing network packets (WIP?)
    - multikey: cryptographic keys and artifacts

### InterPlanetary Naming System (IPNS)?

- Immutable content links are not very useful for sharing, you'd have to distribute a new link every time you change the content.
- IPNS allows you to create a mutable link.
- IPNS link is the hash of a public key
- IPNS is made using a distributed hash table (DHT)
- There is no historical information kept, only the latest mapping.
- IPNS forgets values after about 12 hours, it has a TTL property though but not sure if you can exceed it
- Links that are generated uses a peer-id but it's ugly looking just like IPFS links.

### DNSLink

- This is a separate way to create mutable addresses on IPFS from IPNS
- DNSLink is faster than IPNS and has more reliable names
- Uses DNS TXT records to map a domain name to an IPFS address
- Uses the same type of addressing through the proxy server as IPNS: /ipns/vaibhavsaini.com
- TXT entry looks like: `"dnslink=/ipns/QmcQrDGFUyHdU5seZSNz6XH9dF6sM4A6JXpJAuvKzKZUUP"`

### To host an IPFS site:

- Create a DNS A record pointing our sub-domain to the IP address of an IPFS peer listening on port 80 for HTTP requests 
- Or a CNAME pointing to a gateway so that if the IP changes it still works.


You can generate an IPNS name for a CID via: 

```ipfs name publish QmcQrDGFUyHdU5seZSNz6XH9dF6sM4A6JXpJAuvKzKZUUP```

More specifically using a specific key:
ipfs key gen --type=rsa --size=2048 my_private_key
ipfs name publish --key=my_private_key <cid_to_my_blog>


You can resolve which CID is linked to a peer-id with:

```ipfs name resolve QmcQrDGFUyHdU5seZSNz6XH9dF6sM4A6JXpJAuvKzKZUUP```

You can issue the command again to update the link.

This command takes a long time, even in the minutes.

> Published to QmVPoGzjXbwQ1eUw4uPpCGkFYZmyrLqZsvdSL8diKDZDkN: /ipfs/QmcQrDGFUyHdU5seZSNz6XH9dF6sM4A6JXpJAuvKzKZUUP


### Libp2p

- communicate IPFS data
- Peer discovery
- Libra uses Libp2p as their network layer
- Identity is provided via a peer-id which is a public key hash
    - you can get the public key from it through a lookup table?
    - To communicate you need a peer-id and their location
    - Peer routing is the process of discovering a location based on a peer-id
        - Answer will come directly if it is known, otherwise peers will look for you and respond
    - Peer routing in libp2p uses a DHT using the  Kademlia routing algorithm. Kademlia slide deck here https://docs.google.com/presentation/d/11qGZlPWu6vEAhA7p3qsQaQtWH7KofEC9dMeBFZ1gYeA/mobilepresent?slide=id.g1718cc2bc_08645
    - Content routing uses the same Kademlia routing algorithm as peer routing.
- Pubsub interface, short for publish subscribe allows for communication with all subscribers of events.
    - https://github.com/libp2p/specs/tree/master/pubsub
    - floodsub, gossipsub, episub
- LibP2P is composed of:
    - Exchange protocols: BitTorrent, Bitswap, FTP, HTTP
    - Routing protocols: Gossip, Chord, Kad DHT, mDNS, Delegated, I2P, Tor
    - Network protocols: CJDNS, UDT, uTP, WebRTP, QUIC, TCP, Websockets, I2P, Tor
- LibP2P is made up of core modules:
    - Content routing, peer routing, discovery, transports, NAT traversal
- Has implementations in all popular languages
- 


You can lookup DNS TXT entry with dig

`dig +noall +answer TXT vaibhavsaini.com`

> vaibhavsaini.com.	3600	IN	TXT	"dnslink=/ipns/Qmb1VVr5xjpXHCTcVm3KF3i88GLFXSetjcxL7PQJRviXSy"

## Filecoin

- Proof of replication
- Uses Proof of storage
- It's a decentralized storage network (DSN)
- Filecoin has 3 types of users:
- Clients
    - Pay to store and retrieve data
- Storage miners
    - Store's client data for a reward
    - They pick how much space they are willing to reserve for storage
    - Must provide proof that he still has the data
    - Everyone can look at the proofs to make sure it is still valid, without the need to store the data
- Retrieval miners
    - Provide client data when requested
    - They get the data from clients or storage miners
    - Data is split into pieces and a micropayments happens for each piece.
- A node can be both a storage and retrieval minder
- Clients submit bid orders, specifying the price they want to pay. Miners submit ask orders, specifying the price they want to receive. A deal order is submitted to the blockchain if they are matched.
- There is a Storage Order Book.
- Clients put Puts on the storage order book.
- Storage miners Pledge their storage by depositing collateral
    - If proof of storage fails, the pledge is lost.
- Each sector that a storage miner offers, is stored into the allocation table
- A sector is said to be "sealed" when a storage sector is filled.
- Retrieval markets is off-chain
- The probability that the network elects a miner to create a new block (we refer to this as the voting power of the miner) is proportional to their storage currently in use in relation to the rest of the network
- Has 2 consensus algorithms: 
    - Proof-of-Replication (PoRep)
    - Proof-of-Spacetime (PoSt)
- There is smart contract support
    - File contracts
    - Smart contracts


## Other resources

Tim Berners-Lee Solid
Read more abourt solid project  decentralized web /linked data
https://solid.mit.edu/

Decentralized protocols school
https://proto.school/#/data-structures/01 (already read)
https://proto.school/#/data-structures/resources <- other tutorials not yet read


IPNS, DNSLink and more: https://hackernoon.com/ten-terrible-attempts-to-make-the-inter-planetary-file-system-human-friendly-e4e95df0c6fa

LibP2P core concepts:
- https://docs.libp2p.io/concepts/?source=post_page-----f8bf7724d452----------------------

Simple as water tutorials:
- https://simpleaswater.com/?ref=medium_multiformats&source=post_page-----cf25eef83966----------------------

- https://simpleaswater.com/?ref=hackernoon_ipfs
