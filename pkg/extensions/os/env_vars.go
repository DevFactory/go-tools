/* Copyright 2019 DevFactory FZ LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package os

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnvVarWithDefault loads value of environment variable varName. If
// the variable is not set, the defaultVal is returned
func GetEnvVarWithDefault(varName, defaultVal string) string {
	if val, found := os.LookupEnv(varName); found {
		return val
	}
	return defaultVal
}

// GetBoolFromEnvVarWithDefault loads environment variable varName and tries to parse
// it as bool. If the environment variable is not set, the defaultVal is returned.
// If the variable was set, but parsing failed, the defaultVal is returned and
// error is set.
func GetBoolFromEnvVarWithDefault(varName string, defaultVal bool) (bool, error) {
	val, err := strconv.ParseBool(GetEnvVarWithDefault(varName, "true"))
	if err == nil {
		return val, nil
	}
	return defaultVal, err
}

// GetIntFromEnvVarWithDefault loads environment variable varName and tries to parse
// it as int. If the environment variable is not set, the defaultVal is returned.
// If the variable was set, but parsing failed, the defaultVal is returned and
// error is set.
func GetIntFromEnvVarWithDefault(varName string, defaultVal int) (int, error) {
	val, err := strconv.ParseInt(GetEnvVarWithDefault(varName, fmt.Sprintf("%d", defaultVal)), 10, 64)
	return int(val), err
}

// GetUIntFromEnvVarWithDefault loads environment variable varName and tries to parse
// it as uint. If the environment variable is not set, the defaultVal is returned.
// If the variable was set, but parsing failed, the defaultVal is returned and
// error is set.
func GetUIntFromEnvVarWithDefault(varName string, defaultVal uint) (uint, error) {
	val, err := strconv.ParseUint(GetEnvVarWithDefault(varName, fmt.Sprintf("%d", defaultVal)), 10, 64)
	return uint(val), err
}

// GetFloatFromEnvVarWithDefault loads environment variable varName and tries to parse
// it as float64. If the environment variable is not set, the defaultVal is returned.
// If the variable was set, but parsing failed, the defaultVal is returned and
// error is set.
func GetFloatFromEnvVarWithDefault(varName string, defaultVal float64) (float64, error) {
	return strconv.ParseFloat(GetEnvVarWithDefault(varName, fmt.Sprintf("%f", defaultVal)), 64)
}
