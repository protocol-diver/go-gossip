package gogossip

type Packet interface {
	SetID([8]byte)
	ID() [8]byte
	Kind() byte
}

const (
	PullReqestType   = 1
	PullResponseType = 2
)

type (
	PullRequest struct {
		id     [8]byte
		Target [8]byte
	}
	PullResponse struct {
		id   [8]byte
		Data []byte
	}
)

func (req *PullRequest) SetID(id [8]byte) { panic("not supported") }
func (req *PullRequest) ID() [8]byte      { return req.id }
func (req *PullRequest) Kind() byte       { return PullReqestType }

func (res *PullResponse) SetID(id [8]byte) { res.id = id }
func (res *PullResponse) ID() [8]byte      { return res.id }
func (res *PullResponse) Kind() byte       { return PullResponseType }
