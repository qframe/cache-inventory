package qcache_inventory

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/qnib/qframe-types"
	"github.com/zpatrick/go-config"
)

func TestNew(t *testing.T) {
	cfg := config.NewConfig([]config.Provider{})
	qChan := qtypes.NewCfgQChan(cfg)
	_, err := New(qChan, cfg, "test")
	assert.NoError(t, err, "should not cause trouble")
}