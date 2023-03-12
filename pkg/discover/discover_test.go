package discover_test

import (
	"context"
	"testing"

	"github.com/kjuulh/shuttletask/pkg/discover"
	"github.com/stretchr/testify/assert"
)

func TestDiscover(t *testing.T) {
	discovered, err := discover.Discover(context.Background(), "testdata/simple/shuttle.yaml")
	assert.NoError(t, err)

	assert.Equal(t, discover.Discovered{
		Local: &discover.ShuttleTaskDiscovered{
			Files: []string{
				"build.go",
				"download.go",
			},
			DirPath:   "testdata/simple/shuttletask",
			ParentDir: "testdata/simple",
		},
	}, *discovered)

}
