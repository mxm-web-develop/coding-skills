# Mobile implementation

- Respect platform lifecycle, navigation, state-management, concurrency, and design-system conventions.
- Separate platform/UI rendering from domain state transitions and data adapters.
- Cover loading, empty, error, offline, permission, rotation/resizing, accessibility, and localization states where applicable.
- Avoid oversized screens/view models; split by feature responsibility and keep effects injectable for tests.
- UI changes require platform functional tests plus screenshot/golden review when the existing stack supports it.
