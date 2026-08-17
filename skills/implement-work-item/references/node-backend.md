# Node.js backend implementation

- Separate transport/controllers, application use cases, domain logic, and infrastructure adapters when the project architecture supports those concepts.
- Validate untrusted data at the boundary and retain typed internal values.
- Keep network, filesystem, database, process, and clock effects behind small interfaces or injected functions where practical.
- Use the repository's module system, runtime target, package manager, and async error conventions.
- Do not create a generic service layer that only forwards calls; name modules after domain behavior.
