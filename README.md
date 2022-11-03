# go-gossip
This is Go implementation of the Gossip protocol.<br>
<br>

Gossip is a communication protocol that delivers messages in a distributed system. <br>
The point is that each node transmits data while periodically exchanging metadata based on TCP/UDP without a broadcast master. <br>
In general, each node periodically performs health checks of other nodes and communicates with them, but this library relies on an externally imported discovery layer. <br>
The gossip protocol is divided into two main categories: Push and Pull. If implemented as a push, it becomes inefficient if a large number of peers are already infected. If implemented as Pull, it will propagate efficiently, but there is a point to be concerned about because the message that needs to be propagated to the peer needs to be managed. <br>
In this project, by properly mixing push and pull methods, it is possible to efficiently propagate even when a large number of peers are already infected, and the goal is to reduce the difficulty of managing messages that need to be propagated to other peers. <br>
it works almost identically to the existing Push-based gossip protocol. It selects a set number of random peers for a new message and send the message. The difference here is that the peer that receives the 'Push' message sends an ACK message to the sender. <br>
If the target peer does not operate normally, or if the message has already received before, does not send ACK message. <br>
The sender collects the number of ACKs to see if it has received a majority of the number of messages it has sent. If a 'Push' message is sent to 3 random peers, the 'Push' process will correctly end only when two or more 'ACK' are received. <br>
What if I didn't get more than a majority? <br>
Suppose you sent a 'Push' message to 3 random peers, but only received 1 'ACK'. The sender adds a certain value to the previously established number of peers 3 and sends, for example, 'PullSync' to 5 peers. The message is sent with the id of the data the sender was trying to propagate. The peer receiving the 'PullSync' sends a 'Pull request' including the data ID to the sender of the message. Finally, the original sender peer sends a 'Pull response' containing the requested data, and then deletes the data from the memory. <br>
If implemented as above, the inefficiency of the push-based gossip protocol, which requires sending and receiving messages between already infected peers, and the hassle of managing data to respond to pull requests can be reduced. <br>

Take a look at the list of supported features below. <br>

- Message propagation
- Secure transport

It's forcus on gossip through relying on discovery layer from outside.


# Layer
<< LAYER IMAGE >> <br><br>
## Discovery layer
Discovery layer serves as the managed peer table. That could be static peer table, also could dynamic peer table(like DHT). <br>
The required(MUST) method is Gossipiers. Gossipiers is used to select random peers to send gossip messages to. <br>

```go
type Discovery interface {
	Gossipiers() []string
}
```
(It means DHT or importers covers registration to Gossip protocol) <br>
Gossipiers returns array of raw addresses. The raw addresses will validate when gossip layer. Even if rely on validation of externally imported methods, We need to double-check internally here(trust will make unexpected panic).<br>
A consideration is whether Gossipiers return value actually needs a peer ID. <br>
In generally, there is need each peer's meta datas(about memberlist), the unique id is required. However, since this library does not support health checks, the peer id is not required from a metadata required point of view. <br>
In addition, if you receive and use a unique ID from outside, the dependency relationship becomes severe, so I think it is correct not to have an peer id. <br>
When making a request, such as checking gossip node stat or something, we decide to use the raw address.

## Gossip layer
Gossip layer serves core features that propagating gossip messages and relay data to application when needed. <br>
For serve that, it's detect packet and handles them correctly to the packet types. <br>
There is three tasks what this layer have to do. <br>

1. The node should be able to be a gossip culprit. Provide surface interface for application programs to push gossip messages.
2. MUST have to respond like, PUSH_MESSAGE -> PUSH_ACK, PULL_SYN -> PULL_REQUEST, PULL_REQUEST -> PULL_RESPONSE.
3. Detects is the gossip message already exist in memory and relay the gossip messages to the application if necessary.

The gossip node needs two buffers: a buffer where the application program can receive gossip messages and a buffer to temporarily store it to propagate to other gossip nodes. <br>
I decided to use the LRU cache for message temporary storage for propagation. Gossip messages that no longer propagate are likely not to be referenced by other nodes in the future. One thing to consider here is that the size of the cache should be much larger than the number of times a node can make PUSH requests. Otherwise, the cache will be replaced as soon as it starts propagate gossip messages. <br>
We need to find a suitable config values. <br>


## Transport/Security layer
Security layer resolve the secure between peer to peer trnasmission. It should be possible to add multiple encryption algorithms. I'm just considering a method of encrypting and decrypting using a passphrase. <br>

Packet<br>
```
┏---------------------┓
| Label | Actual data |
┗---------------------┛
```

Label
```
┏----------------------------------------------------------┓
| Packet type| Encrypt algorithm | Actual data size (May?) | 
┗----------------------------------------------------------┛
```
Packet type (1 byte) <br>
> 1: PushMessage <br>
> 2: PullSync <br>
> 3: PullRequest <br>
> 4: PullResponse <br>

Encrypt alogrithm (1 byte) <br>
> 1: AES-256-CBC <br>

Actual data size (4 byte); BigEndian ordered uint32 <br>
This is not necessary unless you add a specific flag (eg checksum) after the data.

PULL_REQUEST
```
```

PULL_RESPONSE
```
```
