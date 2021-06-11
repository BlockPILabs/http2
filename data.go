package http2

import (
	"sync"

	"github.com/dgrr/http2/http2utils"
)

const FrameData FrameType = 0x0

// Data defines a FrameData
//
// Data frames can have the following flags:
// END_STREAM
// PADDED
//
// https://tools.ietf.org/html/rfc7540#section-6.1
type Data struct {
	endStream  bool
	hasPadding bool
	b          []byte // data bytes
}

var dataPool = sync.Pool{
	New: func() interface{} {
		return &Data{}
	},
}

// AcquireData ...
func AcquireData() (data *Data) {
	data = dataPool.Get().(*Data)
	data.Reset()
	return
}

// ReleaseData ...
func ReleaseData(data *Data) {
	dataPool.Put(data)
}

// Reset ...
func (data *Data) Reset() {
	data.endStream = false
	data.hasPadding = false
	data.b = data.b[:0]
}

// CopyTo copies data to d.
func (data *Data) CopyTo(d *Data) {
	d.hasPadding = data.hasPadding
	d.endStream = data.endStream
	d.b = append(d.b[:0], data.b...)
}

// SetEndStream ...
func (data *Data) SetEndStream(value bool) {
	data.endStream = value
}

func (data *Data) EndStream() bool {
	return data.endStream
}

// Data returns the byte slice of the data readed/to be sendStream.
func (data *Data) Data() []byte {
	return data.b
}

// SetData resets data byte slice and sets b.
func (data *Data) SetData(b []byte) {
	data.b = append(data.b[:0], b...)
}

// Padding returns true if the data will be/was hasPaddingded.
func (data *Data) Padding() bool {
	return data.hasPadding
}

// SetPadding sets hasPaddingding to the data if true. If false the data won't be hasPaddingded.
func (data *Data) SetPadding(value bool) {
	data.hasPadding = value
}

// Append appends b to data
func (data *Data) Append(b []byte) {
	data.b = append(data.b, b...)
}

func (data *Data) Len() uint32 {
	return uint32(len(data.b))
}

// Write writes b to data
func (data *Data) Write(b []byte) (int, error) {
	n := len(b)
	data.Append(b)
	return n, nil
}

// ReadFrame reads data from fr.
//
// This function does not reset the Frame.
func (data *Data) ReadFrame(fr *Frame) (err error) {
	payload := fr.Payload()
	if fr.HasFlag(FlagPadded) {
		payload = http2utils.CutPadding(payload, fr.Len())
	}

	data.endStream = fr.HasFlag(FlagEndStream)
	data.b = append(data.b[:0], payload...)

	return
}

// WriteFrame writes the data to the frame payload setting FlagPadded.
//
// This function only resets the frame payload.
func (data *Data) WriteFrame(fr *Frame) {
	// TODO: generate hasPadding and set to the frame payload
	fr.SetType(FrameData)

	if data.endStream {
		fr.AddFlag(FlagEndStream)
	}

	if data.hasPadding {
		fr.AddFlag(FlagPadded)
		data.b = http2utils.AddPadding(data.b)
	}

	fr.SetPayload(data.b)
}
