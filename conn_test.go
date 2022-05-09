package fastafpacket

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

func TestParseTimestamps(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		hardware := time.Date(1982, 1, 13, 1, 1, 1, 1, time.UTC)
		software := hardware.Add(10 * time.Millisecond)

		msg := unix.SocketControlMessage{
			Header: unix.Cmsghdr{
				Level: unix.SOL_SOCKET,
				Type:  SocketOptTimestamping,
			},
			Data: make([]byte, 48),
		}

		hardwarenano := hardware.UnixNano()
		softwarenano := software.UnixNano()

		binary.LittleEndian.PutUint64(msg.Data[0:], 0)
		binary.LittleEndian.PutUint64(msg.Data[8:], uint64(softwarenano))

		binary.LittleEndian.PutUint64(msg.Data[32:], 0)
		binary.LittleEndian.PutUint64(msg.Data[40:], uint64(hardwarenano))

		ts, err := ParseSocketTimestamps(msg)
		assert.NoError(t, err)

		assert.True(t, ts.Hardware.Equal(hardware))
		assert.True(t, ts.Software.Equal(software))
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		msg := unix.SocketControlMessage{
			Header: unix.Cmsghdr{
				Level: unix.SOL_XDP,
			},
		}

		ts, err := ParseSocketTimestamps(msg)
		assert.Error(t, err)

		assert.True(t, ts.Hardware.IsZero())
		assert.True(t, ts.Software.IsZero())
	})
}
