# Go and Rust tests

For Go, use table-driven subtests when cases share behavior, run targeted package tests first, and add race/fuzz/integration checks when the changed risk requires them. Keep external-resource tests explicit and deterministic.

For Rust, use module/unit tests for core rules, integration tests for public crate behavior, and property/fuzz/compile-fail tests when invariants or APIs warrant them. Run the detected formatter, lints, features, and workspace matrix.

For both, test cancellation, concurrency ownership, error context, serialization, and platform/feature variants affected by the change.
