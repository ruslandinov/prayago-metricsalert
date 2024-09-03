package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	agent := NewAgent(NewAgentConfig())
	agent.updateMetrics()

	assert.NotEmpty(t, agent.metrics)
	assert.NotEmpty(t, agent.randomValue.Value)
	assert.EqualValues(t, agent.pollCount.Delta, 1)
}
