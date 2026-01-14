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

package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestViper_AutomaticEnv_RespectsKeyReplacerSetAfter(t *testing.T) {
	v := viper.New()

	v.SetEnvPrefix("DATAROBOT_CLI")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	t.Setenv("DATAROBOT_CLI_SKIP_AUTH", "true")

	if !v.GetBool("skip-auth") {
		t.Fatalf("expected viper.GetBool(\"skip-auth\") to resolve DATAROBOT_CLI_SKIP_AUTH when SetEnvKeyReplacer is called after AutomaticEnv")
	}

	if !v.GetBool("skip_auth") {
		t.Fatalf("expected viper.GetBool(\"skip_auth\") to resolve DATAROBOT_CLI_SKIP_AUTH when SetEnvKeyReplacer is called after AutomaticEnv")
	}
}

func TestViper_AutomaticEnv_RespectsReverseKeyReplacerSetAfter(t *testing.T) {
	v := viper.New()

	v.SetEnvPrefix("DATAROBOT_CLI")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("_", "-"))

	// This should NOT work: SetEnvKeyReplacer affects how a Viper key is transformed
	// into an ENV var name for lookup. It does not transform the OS env var name.
	// With '_' -> '-', key "skip_auth" maps to env var DATAROBOT_CLI_SKIP-AUTH.
	// But environment variable names with '-' are generally not usable/portable.
	t.Setenv("DATAROBOT_CLI_SKIP_AUTH", "true")

	if v.GetBool("skip_auth") {
		t.Fatalf("expected viper.GetBool(\"skip_auth\") to be false when SetEnvKeyReplacer maps '_' -> '-' because it will look for DATAROBOT_CLI_SKIP-AUTH (not DATAROBOT_CLI_SKIP_AUTH)")
	}

	if v.GetBool("skip-auth") {
		t.Fatalf("expected viper.GetBool(\"skip-auth\") to be false when SetEnvKeyReplacer maps '_' -> '-' because it will look for DATAROBOT_CLI_SKIP-AUTH (not DATAROBOT_CLI_SKIP_AUTH)")
	}
}

func TestViper_AutomaticEnv_DoesNotUseReplacerIfNeverSet(t *testing.T) {
	v := viper.New()

	v.SetEnvPrefix("DATAROBOT_CLI")
	v.AutomaticEnv()

	t.Setenv("DATAROBOT_CLI_SKIP_AUTH", "true")

	if v.GetBool("skip-auth") {
		t.Fatalf("expected viper.GetBool(\"skip-auth\") to be false without SetEnvKeyReplacer; replacer is required to map '-' to '_' for env var lookup")
	}

	if !v.GetBool("skip_auth") {
		t.Fatalf("expected viper.GetBool(\"skip_auth\") to resolve DATAROBOT_CLI_SKIP_AUTH without SetEnvKeyReplacer")
	}
}

func TestViper_AutomaticEnv_RespectsReplacerSetBefore(t *testing.T) {
	v := viper.New()

	v.SetEnvPrefix("DATAROBOT_CLI")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	t.Setenv("DATAROBOT_CLI_SKIP_AUTH", "true")

	if !v.GetBool("skip-auth") {
		t.Fatalf("expected viper.GetBool(\"skip-auth\") to resolve DATAROBOT_CLI_SKIP_AUTH when SetEnvKeyReplacer is called before AutomaticEnv")
	}

	if !v.GetBool("skip_auth") {
		t.Fatalf("expected viper.GetBool(\"skip_auth\") to resolve DATAROBOT_CLI_SKIP_AUTH when SetEnvKeyReplacer is called before AutomaticEnv")
	}
}

func TestViper_AutomaticEnv_NoPrefix(t *testing.T) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	key := "SKIP_AUTH"
	os.Setenv(key, "true")

	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})

	if !v.GetBool("skip-auth") {
		t.Fatalf("expected viper.GetBool(\"skip-auth\") to resolve SKIP_AUTH when no prefix is set")
	}

	if !v.GetBool("skip_auth") {
		t.Fatalf("expected viper.GetBool(\"skip_auth\") to resolve SKIP_AUTH when no prefix is set")
	}
}
