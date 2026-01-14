// Copyright 2025 DataRobot, Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package start

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepCompleteMsg_String(t *testing.T) {
	tests := []struct {
		name     string
		msg      stepCompleteMsg
		expected string
	}{
		{
			name:     "empty message",
			msg:      stepCompleteMsg{},
			expected: `stepCompleteMsg{message: "", waiting: false, done: false, hideMenu: false, quickstartScriptPath: "", selfUpdate: false, executeScript: false, needTemplateSetup: false}`,
		},
		{
			name: "message with text",
			msg: stepCompleteMsg{
				message: "Test message",
			},
			expected: `stepCompleteMsg{message: "Test message", waiting: false, done: false, hideMenu: false, quickstartScriptPath: "", selfUpdate: false, executeScript: false, needTemplateSetup: false}`,
		},
		{
			name: "all boolean flags set",
			msg: stepCompleteMsg{
				waiting:           true,
				done:              true,
				hideMenu:          true,
				selfUpdate:        true,
				executeScript:     true,
				needTemplateSetup: true,
			},
			expected: `stepCompleteMsg{message: "", waiting: true, done: true, hideMenu: true, quickstartScriptPath: "", selfUpdate: true, executeScript: true, needTemplateSetup: true}`,
		},
		{
			name: "with quickstart script path",
			msg: stepCompleteMsg{
				quickstartScriptPath: "/path/to/quickstart.sh",
			},
			expected: `stepCompleteMsg{message: "", waiting: false, done: false, hideMenu: false, quickstartScriptPath: "/path/to/quickstart.sh", selfUpdate: false, executeScript: false, needTemplateSetup: false}`,
		},
		{
			name: "complete example with all fields",
			msg: stepCompleteMsg{
				message:              "Script found",
				waiting:              true,
				done:                 false,
				hideMenu:             false,
				quickstartScriptPath: "./quickstart.sh",
				selfUpdate:           false,
				executeScript:        true,
				needTemplateSetup:    false,
			},
			expected: `stepCompleteMsg{message: "Script found", waiting: true, done: false, hideMenu: false, quickstartScriptPath: "./quickstart.sh", selfUpdate: false, executeScript: true, needTemplateSetup: false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.msg.String()
			assert.Equal(t, tt.expected, result, "String() output should match expected format")
		})
	}
}
