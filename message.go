package gogossip

type Packet interface {
	ID() uint32
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
		id   uint32
		Key  [8]byte
		Data []byte
	}
	PushAck struct {
		id  uint32
		Key [8]byte
	}
	PullSync struct {
		id     uint32
		Target [8]byte
	}
	PullRequest struct {
		id     uint32
		Target [8]byte
	}
	PullResponse struct {
		id     uint32
		Target [8]byte
		Data   []byte
	}
)

func (p *PushMessage) ID() uint32 { return p.id }
func (p *PushMessage) Kind() byte { return PullRequestType }

func (p *PushAck) ID() uint32 { return p.id }
func (p *PushAck) Kind() byte { return PullRequestType }

func (p *PullSync) ID() uint32 { return p.id }
func (p *PullSync) Kind() byte { return PullRequestType }

func (req *PullRequest) ID() uint32 { return req.id }
func (req *PullRequest) Kind() byte { return PullRequestType }

func (res *PullResponse) ID() uint32 { return res.id }
func (res *PullResponse) Kind() byte { return PullResponseType }
