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
	PullRequest struct {
		id uint32
	}
	PullResponse struct {
		id     uint32
		Keys   [][8]byte
		Values [][]byte
	}
)

func (req *PullRequest) ID() uint32 { return req.id }
func (req *PullRequest) Kind() byte { return PullRequestType }

func (res *PullResponse) ID() uint32 { return res.id }
func (res *PullResponse) Kind() byte { return PullResponseType }
