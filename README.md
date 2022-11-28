# go-gossip
This is Go implementation of the Gossip protocol.<br>


Gossip is a communication protocol that delivers messages in a distributed system. <br>
The point is that each node transmits data while periodically exchanging metadata based on TCP/UDP without a broadcast master. <br>
In general, each node periodically performs health checks of other nodes and communicates with them, but this library relies on an externally imported discovery layer. <br>
The gossip protocol is divided into two main categories: Push and Pull. If implemented as a push, it becomes inefficient if a large number of peers are already infected. If implemented as Pull, it will propagate efficiently, but there is a point to be concerned about because the message that needs to be propagated to the peer needs to be managed. <br>
This project implements the Pull-based Gossip protocol. That's why we need to implement a way to send a new message when another node requests it. <br>
We could use a way to store the message in memory and increment the count when another node requests it, but here we put a timeout on the message for  cleaned up. <br>

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
2. Handles them correctly to the packet types.
3. Detects is the gossip message already exist in memory and relay the gossip messages to the application if necessary.

Take a look the packet specification below. <br>

Packet<br>
```
┏---------------------┓
| Label | Actual data |
┗---------------------┛
```

Label
```
┏--------------------------------┓
| Packet type| Encrypt algorithm | 
┗--------------------------------┛
```
Packet type (1 byte) <br>
> 1: PullRequest <br>
> 2: PullResponse <br>

### Packet handle
<b>PullRequest</b> - It replies a packet to the requester. However, packets that have already been taken by the requester are excluded. <br>
<b>PullResponse</b> - Stores the received message in memory. Messages that have already been received will be ignored. <br>


## Transport/Security layer
Transport layer supports peer-to-peer TCP communication. UDP is sufficient If the communication is simply sending a PullRequest to a random peer. However, this library's PullResponse payload size is unpredictable and will likely exceed the safe from packet fragmentation UDP size of 508 bytes(The minimum size of IPv4 datagram every device has to be able to receive(i.g. MTU) is 576 byte. And IP header(with fully option field) is 60 byte, UDP header is 8 byte. So I guess safety maximum UDP packet size from packet fragmentation is 508(576 - 60 - 8) byte). Implementing with UDP requires chopping packets at the application level and sending them over UDP to avoid packet fragmentation. There's no basis for doing that. <br>
Security layer resolve the secure between peer to peer trnasmission. It should be possible to add multiple encryption algorithms. I'm just considering a method of encrypting and decrypting using a passphrase(It is also okay to encrypt in the application and then propagate it. In this case, you should set NO-SECURE in config). <br>

Encrypt alogrithm (1 byte) <br>
> 0: NO-SECURE <br>
> 1: AES-256-CBC <br>
