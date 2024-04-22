//go:build unit
// +build unit

// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package restart

import (
	"testing"
	"time"

	apicontainerstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/container/status"

	"github.com/stretchr/testify/assert"
)

func TestShouldRestart(t *testing.T) {
	rt := NewRestartTracker(RestartPolicy{Enabled: false, IgnoredExitCodes: []int{0}, AttemptResetPeriod: 1 * time.Minute})
	testCases := []struct {
		name           string
		rp             RestartPolicy
		exitCode       int
		startedAt      time.Time
		desiredStatus  apicontainerstatus.ContainerStatus
		expected       bool
		expectedReason string
	}{
		{
			name:           "restart policy disabled",
			rp:             RestartPolicy{Enabled: false, IgnoredExitCodes: []int{0}, AttemptResetPeriod: 1 * time.Minute},
			exitCode:       1,
			startedAt:      time.Now().Add(2 * time.Minute),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       false,
			expectedReason: "restart policy is not enabled",
		},
		{
			name:           "ignored exit code",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: time.Minute},
			exitCode:       0,
			startedAt:      time.Now().Add(2 * time.Minute),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       false,
			expectedReason: "exit code 0 should be ignored",
		},
		{
			name:           "non ignored exit code",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: 1 * time.Minute},
			exitCode:       1,
			startedAt:      time.Now().Add(-2 * time.Minute),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       true,
			expectedReason: "",
		},
		{
			name:           "nil exit code",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: 1 * time.Minute},
			exitCode:       -1,
			startedAt:      time.Now().Add(2 * time.Minute),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       false,
			expectedReason: "exit code is nil",
		},
		{
			name:           "desired status stopped",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: time.Minute},
			exitCode:       1,
			startedAt:      time.Now().Add(2 * time.Minute),
			desiredStatus:  apicontainerstatus.ContainerStopped,
			expected:       false,
			expectedReason: "container's desired status is stopped",
		},
		{
			name:           "attempt reset period not elapsed",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: time.Minute},
			exitCode:       1,
			startedAt:      time.Now(),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       false,
			expectedReason: "attempt reset period has not elapsed",
		},
		{
			name:           "attempt reset period not elapsed within one second",
			rp:             RestartPolicy{Enabled: true, IgnoredExitCodes: []int{0}, AttemptResetPeriod: time.Minute},
			exitCode:       1,
			startedAt:      time.Now().Add(-time.Second * 59),
			desiredStatus:  apicontainerstatus.ContainerRunning,
			expected:       false,
			expectedReason: "attempt reset period has not elapsed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rt.restartPolicy = tc.rp

			// Because we cannot instantiate int pointers directly,
			// check for the exit code and leave this int pointer as nil
			// if there is no value to override it.
			var exitCodeAdjusted *int
			if tc.exitCode != -1 {
				exitCodeAdjusted = &tc.exitCode
			}

			shouldRestart, reason := rt.ShouldRestart(exitCodeAdjusted, tc.startedAt, tc.desiredStatus)
			assert.Equal(t, tc.expected, shouldRestart)
			assert.Equal(t, tc.expectedReason, reason)
		})
	}
}

func TestRecordRestart(t *testing.T) {
	rt := NewRestartTracker(RestartPolicy{Enabled: false, IgnoredExitCodes: []int{0}, AttemptResetPeriod: 1 * time.Minute})
	assert.Equal(t, 0, rt.RestartCount)
	for i := 1; i < 1000; i++ {
		rt.RecordRestart()
		assert.Equal(t, i, rt.RestartCount)
	}
}

func TestRecordRestartPolicy(t *testing.T) {
	rt := NewRestartTracker(RestartPolicy{Enabled: false, AttemptResetPeriod: 1 * time.Minute})
	assert.Equal(t, 0, rt.RestartCount)
	assert.Equal(t, nil, rt.restartPolicy)
}
