# JVM and .NET implementation

- Respect the detected framework's dependency injection, lifecycle, configuration, async, and error conventions.
- Keep controllers/endpoints thin; place domain decisions in framework-light classes or functions and isolate persistence/integration adapters.
- Prefer immutable values and explicit state transitions where supported.
- Split multipurpose classes by responsibility; avoid generic `Manager`, `Helper`, or `Common` containers.
- Preserve binary/public API compatibility and serialization behavior unless the accepted decision permits change.
