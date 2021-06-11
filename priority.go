package http2

import (
	"sync"

	"github.com/dgrr/http2/http2utils"
)

const FramePriority FrameType = 0x2

// Priority represents the Priority frame.
//
// https://tools.ietf.org/html/rfc7540#section-6.3
type Priority struct {
	stream uint32
	weight byte
}

var priorityPool = sync.Pool{
	New: func() interface{} {
		return &Priority{}
	},
}

// AcquirePriority gets priority structure from pool.
func AcquirePriority() *Priority {
	pry := priorityPool.Get().(*Priority)
	return pry
}

// ReleasePriority retusn pry to the Priority frame pool.
func ReleasePriority(pry *Priority) {
	pry.Reset()
	priorityPool.Put(pry)
}

// Reset resets priority fields.
func (pry *Priority) Reset() {
	pry.stream = 0
	pry.weight = 0
}

// CopyTo ...
func (pry *Priority) CopyTo(p *Priority) {
	p.stream = pry.stream
	p.weight = pry.weight
}

// Stream returns the Priority frame stream.
func (pry *Priority) Stream() uint32 {
	return pry.stream
}

// SetStream sets the Priority frame stream.
func (pry *Priority) SetStream(stream uint32) {
	pry.stream = stream & (1<<31 - 1)
}

// Weight returns the Priority frame weight.
func (pry *Priority) Weight() byte {
	return pry.weight
}

// SetWeight sets the Priority frame weight.
func (pry *Priority) SetWeight(w byte) {
	pry.weight = w
}

// ReadFrame reads frame payload and decodes the values into pry.
func (pry *Priority) ReadFrame(fr *Frame) (err error) {
	if len(fr.payload) < 5 {
		err = ErrMissingBytes
	} else {
		pry.stream = http2utils.BytesToUint32(fr.payload) & (1<<31 - 1)
		pry.weight = fr.payload[4]
	}
	return
}

// WriteFrame writes pry to the Freame. The Frame payload is resetted.
func (pry *Priority) WriteFrame(fr *Frame) {
	fr.SetType(FramePriority)
	fr.payload = http2utils.AppendUint32Bytes(fr.payload[:0], pry.stream)
	fr.payload = append(fr.payload, pry.weight)
}
