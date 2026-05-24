// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package internal

// FormatFlag is the resolved output format for the current invocation,
// one of: "" (default/styled), "tsv", "json", "yaml".
//
// Storage lives in this leaf package so both cmd/* (which sets it via
// cobra) and views/* (which branches on it during rendering) can read
// it without creating an import cycle.
var FormatFlag string

// IsStructuredOutput reports whether FormatFlag selects a fully-serialized
// format (json/yaml) handled by cmd/common.NewFormatter. TSV is *not*
// structured here: TSV rendering lives in the view layer where curated
// column data is available.
func IsStructuredOutput() bool {
	return FormatFlag == "json" || FormatFlag == "yaml"
}
