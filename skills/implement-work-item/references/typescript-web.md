# TypeScript web implementation

- Respect the detected router, rendering mode, state/data layer, design system, and server/client boundary.
- Organize feature code by responsibility: view components, domain/state logic, data adapters, schemas/types, and tests. Split a component when rendering, data access, and business rules become entangled.
- Keep reusable domain calculations in framework-independent TypeScript functions. Keep effects in hooks, actions, loaders, stores, or adapters appropriate to the framework.
- Preserve accessibility semantics, keyboard behavior, focus, responsive layout, loading, empty, error, and reduced-motion states.
- Avoid premature memoization and duplicated derived state. Follow installed framework/vendor Skills selected by the engineering profile.
- UI changes require the web functional and visual test playbook before completion.
