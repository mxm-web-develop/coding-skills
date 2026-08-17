# Python implementation

- Follow the detected package layout, supported Python version, formatter, linter, and type checker.
- Keep domain calculations in small typed functions; isolate filesystem, network, database, environment, and clock access.
- Prefer explicit data models and dependency parameters over mutable module globals.
- Split packages by domain capability and adapter responsibility. Avoid an ever-growing `utils.py` or `services.py`.
- Preserve sync/async boundaries and existing exception taxonomy. Add context without obscuring the original cause.
