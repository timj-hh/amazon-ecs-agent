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
	"fmt"
	"time"

	apicontainerstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/container/status"
)

// RestartPolicy represents a policy that contains key information considered when
// deciding whether or not a container should be restarted after it has exited.
type RestartPolicy struct {
	Enabled            bool          `json:"enabled"`
	IgnoredExitCodes   []int         `json:"ignoredExitCodes"`
	AttemptResetPeriod time.Duration `json:"attemptResetPeriod"`
}

type RestartTracker struct {
	RestartCount  int `json:"restartCount,omitempty"`
	restartPolicy RestartPolicy
}

func NewRestartTracker(restartPolicy RestartPolicy) *RestartTracker {
	return &RestartTracker{
		restartPolicy: restartPolicy,
	}
}

func (rt *RestartTracker) GetRestartCount() int {
	return rt.RestartCount
}

// RecordRestart updates the restart tracker's metadata after a restart has occurred.
// This metadata is used to calculate when restarts should occur and track how many
// have occurred. It is not the job of this method to determine if a restart should
// occur or restart the container. It is expected to receive a startedAt time from the container runtime.
func (rt *RestartTracker) RecordRestart() {
	rt.RestartCount += 1
}

// ShouldRestart returns whether the container should restart and a reason string
// explaining why not.
func (rt *RestartTracker) ShouldRestart(exitCode *int, startedAt time.Time,
	desiredStatus apicontainerstatus.ContainerStatus) (bool, string) {
	if !rt.restartPolicy.Enabled {
		return false, "restart policy is not enabled"
	}
	if desiredStatus == apicontainerstatus.ContainerStopped {
		return false, "container's desired status is stopped"
	}
	if exitCode == nil {
		return false, "exit code is nil"
	}
	for _, ignoredCode := range rt.restartPolicy.IgnoredExitCodes {
		if ignoredCode == *exitCode {
			return false, fmt.Sprintf("exit code %d should be ignored", *exitCode)
		}
	}
	if time.Since(startedAt) < rt.restartPolicy.AttemptResetPeriod {
		return false, "attempt reset period has not elapsed"
	}

	return true, ""
}
