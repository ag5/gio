// SPDX-License-Identifier: Unlicense OR MIT

package app

import (
	"testing"

	"gioui.org/unit"
)

func TestPresentationTransaction(t *testing.T) {
	var config Config
	if config.disablePresentationTransaction {
		t.Fatal("presentation transactions should be enabled by default")
	}

	PresentationTransaction(false)(unit.Metric{}, &config)
	if !config.disablePresentationTransaction {
		t.Fatal("PresentationTransaction(false) did not disable presentation transactions")
	}

	PresentationTransaction(true)(unit.Metric{}, &config)
	if config.disablePresentationTransaction {
		t.Fatal("PresentationTransaction(true) did not enable presentation transactions")
	}
}
