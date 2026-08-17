# Go implementation

- Keep packages cohesive and named for the capability they provide, not generic layers.
- Prefer small interfaces declared near the consumer. Accept interfaces and return concrete types unless project conventions differ.
- Make core transformations ordinary deterministic functions; pass context and effects through explicit boundaries.
- Wrap errors with useful context, preserve cancellation, and avoid hidden goroutine ownership.
- Split files by responsibility inside a package when a file accumulates unrelated concerns; avoid unnecessary package fragmentation.
