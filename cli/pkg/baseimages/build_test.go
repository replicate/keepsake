package baseimages

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type CountBuilder struct {
	count uint64
}

func (b *CountBuilder) Build(dir string, image string) error {
	atomic.AddUint64(&b.count, 1)
	return nil
}
func (b *CountBuilder) ImageExists(image string) bool { return false }
func (b *CountBuilder) Verbose() bool                 { return false }

func TestCountBaseImages(t *testing.T) {
	b := &CountBuilder{}
	err := BuildBaseImages(b, "project", "registry", "0.1", 0)
	require.NoError(t, err)
	require.Equal(t, 391, int(b.count))
}
