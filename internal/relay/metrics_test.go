package relay

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/bestruirui/octopus/internal/op"
	"github.com/looplj/axonhub/llm"
	"github.com/stretchr/testify/require"
)

func TestApplyAttemptMetricRecordsSelectedModelOnSuccess(t *testing.T) {
	if isolateStatsTest(t) {
		return
	}

	// Given: a successful upstream call selected a model whose name differs from the client group.
	const groupName = "leaderboard-success-group"
	const selectedModel = "leaderboard-success-selected-model"
	item := model.GroupItem{ModelName: selectedModel}
	channel := &model.Channel{ID: -1, Name: "leaderboard-success-channel"}

	// When: the attempt metrics are recorded.
	applyAttemptMetric(item, channel, model.StatsMetrics{InputToken: 10, OutputToken: 5}, time.Second, true)

	// Then: the leaderboard row is keyed by the selected model, never the group alias.
	statsByName := make(map[string]model.StatsModel)
	for _, stats := range op.StatsModelList() {
		statsByName[stats.Name] = stats
	}
	require.Contains(t, statsByName, selectedModel)
	require.NotContains(t, statsByName, groupName)
	require.EqualValues(t, 1, statsByName[selectedModel].RequestSuccess)
}

func isolateStatsTest(t *testing.T) bool {
	t.Helper()
	const childEnv = "OCTOPUS_RELAY_STATS_TEST_CHILD"
	if os.Getenv(childEnv) == t.Name() {
		return false
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^"+regexp.QuoteMeta(t.Name())+"$", "-test.v")
	cmd.Env = append(os.Environ(), childEnv+"="+t.Name())
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	return true
}

func TestHandleAttemptFailureRecordsSelectedModel(t *testing.T) {
	if isolateStatsTest(t) {
		return
	}

	// Given: a client model group whose selected upstream model has a different name.
	const groupName = "leaderboard-test-group"
	const selectedModel = "leaderboard-test-selected-model"
	upstreamErr := errors.New("upstream failed")
	execution := execution{
		request: relayRequest{model: groupName},
		log: LogRecord{LogOverview: LogOverview{
			ID:        requestIDs.Add(1),
			StartedAt: time.Now(),
		}},
	}
	item := model.GroupItem{ModelName: selectedModel}
	channel := &model.Channel{ID: -1, Name: "leaderboard-test-channel"}
	attempt := &LogAttempt{Index: 1, ChannelName: channel.Name, ModelName: selectedModel}
	result := upstreamResult{
		err:   upstreamErr,
		usage: &llm.Usage{PromptTokens: 10, CompletionTokens: 5},
	}

	// When: the real upstream attempt fails and its metrics are recorded.
	done, err := execution.handleAttemptFailure(context.Background(), item, channel, attempt, result, time.Second, false)

	// Then: the leaderboard row is keyed by the selected model, never the group alias.
	require.False(t, done)
	require.ErrorIs(t, err, upstreamErr)
	statsByName := make(map[string]model.StatsModel)
	for _, stats := range op.StatsModelList() {
		statsByName[stats.Name] = stats
	}
	require.Contains(t, statsByName, selectedModel)
	require.NotContains(t, statsByName, groupName)
	require.EqualValues(t, 1, statsByName[selectedModel].RequestFailed)
}
