//go:build !linux

package agent

import "context"

// Other platforms retain periodic reconciliation without a native watcher.
func watchInputDirectories(context.Context, []InputConfig) (<-chan struct{}, func()) {
	return make(chan struct{}), func() {}
}
