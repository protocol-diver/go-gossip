package gogossip

type Packet interface {
	ID() uint32
	Kind() byte
}

const (
	PullRequestType  = 0x01
	PullResponseType = 0x02
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
