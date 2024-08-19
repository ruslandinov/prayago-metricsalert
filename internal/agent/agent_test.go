package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	agent := NewAgent(NewAgentConfig())
	agent.updateMetrics()

	assert.NotEmpty(t, agent.metrics.list)
	assert.NotEmpty(t, agent.metrics.randomValue)
	assert.EqualValues(t, agent.metrics.pollCount, 1)
}
