package gogossip

type Packet interface {
	SetID([8]byte)
	ID() [8]byte
	Kind() byte
}

const (
	PushMessageType  = 0x01
	PushAckType      = 0x02
	PullSyncType     = 0x03
	PullRequestType  = 0x04
	PullResponseType = 0x05
)

type (
	PushMessage struct {
		id   [8]byte
		Data []byte
	}
	PushAck struct {
		id [8]byte
	}
	PullSync struct {
		id     [8]byte
		Target uint
	}
	PullRequest struct {
		id     [8]byte
		Target uint
	}
	PullResponse struct {
		id   [8]byte
		Data []byte
	}
)

func (req *PullRequest) SetID(id [8]byte) { panic("not supported") }
func (req *PullRequest) ID() [8]byte      { return req.id }
func (req *PullRequest) Kind() byte       { return PullRequestType }

func (res *PullResponse) SetID(id [8]byte) { res.id = id }
func (res *PullResponse) ID() [8]byte      { return res.id }
func (res *PullResponse) Kind() byte       { return PullResponseType }
