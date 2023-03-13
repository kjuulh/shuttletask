package executer_test

import (
	"context"
	"log"
	"os/exec"
	"testing"

	"github.com/kjuulh/shuttletask/pkg/executer"
	"github.com/stretchr/testify/assert"
)

func TestRunVersion(t *testing.T) {
	updateShuttle(t, "testdata/child")
	ctx := context.Background()

	err := executer.Run(ctx, "testdata/child/shuttle.yaml", "run", "version")
	assert.NoError(t, err)
}

func updateShuttle(t *testing.T, path string) {
	shuttleCmd := exec.Command("shuttle", "ls")
	shuttleCmd.Dir = path
	if output, err := shuttleCmd.CombinedOutput(); err != nil {
		log.Printf("%s\n", string(output))
		assert.Error(t, err)
	}
}
