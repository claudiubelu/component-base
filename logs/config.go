/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logs

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/config"
	"k8s.io/klog/v2"
)

// Supported klog formats
const (
	DefaultLogFormat = "text"
	JSONLogFormat    = "json"
)

// LogRegistry is new init LogFormatRegistry struct
var LogRegistry = NewLogFormatRegistry()

// loggingFlags captures the state of the logging flags, in particular their default value
// before flag parsing. It is used by UnsupportedLoggingFlags.
var loggingFlags pflag.FlagSet

func init() {
	// Text format is default klog format
	LogRegistry.Register(DefaultLogFormat, nil)

	var fs flag.FlagSet
	klog.InitFlags(&fs)
	loggingFlags.AddGoFlagSet(&fs)
}

// List of logs (k8s.io/klog + k8s.io/component-base/logs) flags supported by all logging formats
var supportedLogsFlags = map[string]struct{}{
	"v": {},
	// TODO: support vmodule after 1.19 Alpha
}

// BindLoggingFlags binds the Options struct fields to a flagset
func BindLoggingFlags(c *config.LoggingConfiguration, fs *pflag.FlagSet) {
	// The help text is generated assuming that flags will eventually use
	// hyphens, even if currently no normalization function is set for the
	// flag set yet.
	unsupportedFlags := strings.Join(unsupportedLoggingFlagNames(cliflag.WordSepNormalizeFunc), ", ")
	formats := fmt.Sprintf(`"%s"`, strings.Join(LogRegistry.List(), `", "`))
	fs.StringVar(&c.Format, "logging-format", c.Format, fmt.Sprintf("Sets the log format. Permitted formats: %s.\nNon-default formats don't honor these flags: %s.\nNon-default choices are currently alpha and subject to change without warning.", formats, unsupportedFlags))
	// No new log formats should be added after generation is of flag options
	LogRegistry.Freeze()
	fs.BoolVar(&c.Sanitization, "experimental-logging-sanitization", c.Sanitization, `[Experimental] When enabled prevents logging of fields tagged as sensitive (passwords, keys, tokens).
Runtime log sanitization may introduce significant computation overhead and therefore should not be enabled in production.`)
}

// UnsupportedLoggingFlags lists unsupported logging flags. The normalize
// function is optional.
func UnsupportedLoggingFlags(normalizeFunc func(f *pflag.FlagSet, name string) pflag.NormalizedName) []*pflag.Flag {
	// k8s.io/component-base/logs and klog flags
	pfs := &pflag.FlagSet{}
	loggingFlags.VisitAll(func(flag *pflag.Flag) {
		if _, found := supportedLogsFlags[flag.Name]; !found {
			// Normalization changes flag.Name, so make a copy.
			clone := *flag
			pfs.AddFlag(&clone)
		}
	})

	// Apply normalization.
	pfs.SetNormalizeFunc(normalizeFunc)

	var allFlags []*pflag.Flag
	pfs.VisitAll(func(flag *pflag.Flag) {
		allFlags = append(allFlags, flag)
	})
	return allFlags
}

// unsupportedLoggingFlagNames lists unsupported logging flags by name, with
// optional normalization and sorted.
func unsupportedLoggingFlagNames(normalizeFunc func(f *pflag.FlagSet, name string) pflag.NormalizedName) []string {
	unsupportedFlags := UnsupportedLoggingFlags(normalizeFunc)
	names := make([]string, 0, len(unsupportedFlags))
	for _, f := range unsupportedFlags {
		names = append(names, "--"+f.Name)
	}
	sort.Strings(names)
	return names
}
