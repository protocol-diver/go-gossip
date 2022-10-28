# go-gossip
This is Go implementation of the Gossip protocol.<br>
<br>

Gossip is a communication protocol that delivers messages in a distributed system. <br>
The point is that each node transmits data while periodically exchanging metadata based on TCP/UDP without a broadcast master. <br>
In general, each node periodically performs health checks of other nodes and communicates with them, but this library relies on an externally imported discovery layer.<br>
Take a look at the list of supported features below. <br>

- Message propagation (Pull)
- Secure transport

It's forcus on gossip through relying on discovery layer from outside.


### pseudo code (temporary)
```
loop
	(taskType, task) <- taskQueue
	if taskType is 'PUSH' then
		id, msg = task
		if load(id) is 'empty' then
			save(id)
			relay(application program, msg)
			peer <- randomPeers
			send(peer, PULL_REQ, msg)
		end if
	endif
	
	if taskType is 'PULL_REQ' then
		sender, id = task
		send(sender, UPDATE_RES, load(id))
	end if

	if taskType is 'PULL_RES' then
		id, msg = task
		taskQueue <- ('PUSH', task(id, msg)
	end if
end loop
```


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
2. Service propagation protocol by 'pull'. MUST have to send PULL_REQUEST to randomly selected peers if the node receive newly gossip request. When receive the PULL_REQUEST, the node reply to the sender with the requested value from PULL_REQUEST.
3. Detects is the gossip message already exist in storage and relay the gossip messages to the application if necessary.

The gossip node needs two buffers: a buffer where the application program can receive gossip messages and a buffer to temporarily store it to propagate to other gossip nodes. <br>
I decided to use the LRU cache for message temporary storage for propagation. Gossip messages that no longer propagate are likely not to be referenced by other nodes in the future. One thing to consider here is that the size of the cache should be much larger than the number of times a node can make PUSH requests. Otherwise, the cache will be replaced as soon as it starts propagate gossip messages. <br>
We need to find a suitable config values. <br>


## Transport/Security layer
Security layer resolve the secure between peer to peer trnasmission. It should be possible to add multiple encryption algorithms. The method of sharing the session key is not in mind. I'm just considering a method of encrypting and decrypting using a passphrase. <br>

TEMP <br>

Packet<br>
```
┏---------------------┓
| Label | Actual data |
┗---------------------┛
```

Label
```
┏--------------------------------------┓
| Encrypt algorithm | Actual data size | 
┗--------------------------------------┛
```
Encrypt alogrithm (1 byte)
- 1: RSA-4096 (example)
Actual data size (n byte == maybe max gossip message size?)

PULL_REQUEST
```
```

PULL_RESPONSE
```
```
