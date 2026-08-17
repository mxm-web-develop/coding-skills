# Reviewed community Skill sources

This catalog is a discovery shortlist, not an automatic dependency list. Confirm the current repository contents, license, release/commit, and individual Skill instructions before installation.

| Source | Useful matches | Default trust |
|---|---|---|
| [Vercel Agent Skills](https://github.com/vercel-labs/agent-skills) | React/Next.js performance, React composition, React Native | vendor |
| [Anthropic Skills](https://github.com/anthropics/skills) | frontend design, web application testing, web artifacts | vendor |
| [GitHub Awesome Copilot](https://github.com/github/awesome-copilot) | Jest, pytest, JUnit/Spring, .NET test runners, Vue/Pinia, Playwright workflows | official-ecosystem |
| [Superpowers](https://github.com/obra/superpowers) | TDD, systematic debugging, verification-before-completion, planning/review methodology | established-community |

Selection examples:

- React/Next.js implementation: prefer an installed Vercel React Skill matching the detected version and task.
- Python testing: use an installed pytest-focused Skill only when pytest is the repository's configured runner.
- Java/Spring or .NET: match the exact framework and test runner rather than a generic language Skill.
- Browser testing: combine an installed Playwright/web-testing Skill with AI Flow's evidence and visual-review contract.
- TDD or debugging: methodology Skills may guide the phase but do not replace Work Items, Checkpoints, Evidence, or approval gates.

Do not treat an `awesome` catalog entry as reviewed merely because it is listed. Pin and record the actual selected Skill source.
