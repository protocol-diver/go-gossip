package gogossip

type Packet interface {
	Kind() byte
}

const (
	PullReqestType   = 1
	PullResponseType = 2
)

type (
	PullRequest  struct{}
	PullResponse struct{}
)

func (req *PullRequest) Kind() byte { return PullReqestType }

func (res *PullResponse) Kind() byte { return PullResponseType }
