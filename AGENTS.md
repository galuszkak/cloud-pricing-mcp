### For Project Context

To understand the project, please refer to the following documents:

- **`ARCHITECTURE.md`**: For a detailed overview of the software architecture.
- **`PRD.md`**: For the Product Requirements Document, which outlines the features and functionality of the project.

### Golang Development Guidelines

As a Golang expert, you are expected to adhere to the following best practices to maintain code quality, readability, and performance.

#### 1. Code Formatting
Always format your code using `gofmt` before committing. This ensures a consistent code style across the entire project.

#### 2. Error Handling
- Use the standard `if err != nil` pattern for error handling.
- Errors are values. Return them from functions instead of panicking.
- Panics should only be used for unrecoverable errors that should halt the program.

#### 3. Package Management
This project uses Go Modules. Use `go mod tidy` to keep the `go.mod` and `go.sum` files up-to-date with your changes.

#### 4. Testing
- **Employ a Test-Driven Development (TDD) approach.** Start by writing a failing test that captures the requirements of a new feature before writing the implementation code.
- Suggest comprehensive testing strategies rather than just example tests, including considerations for mocking, test organization, and coverage.
- Write unit tests for your code using the built-in `testing` package.
- Utilize table-driven tests to cover multiple cases efficiently.
- Ensure all new features have corresponding tests and that all tests pass before submitting your changes.

#### 5. Simplicity and Readability
- Favor elegant, maintainable solutions over verbose code. Assume understanding of language idioms and design patterns.
- Write simple, clear, and readable code. Avoid unnecessary complexity and cleverness.
- Follow the principle of "clear is better than clever."

#### 6. Naming Conventions
- Follow Go's idiomatic naming conventions.
- Use `PascalCase` for exported identifiers and `camelCase` for internal ones.
- Keep names short but descriptive.

#### 7. Documentation
- Focus comments on 'why' not 'what' - assume code readability through well-named functions and variables.
- Document all exported packages, functions, types, and variables using `godoc` conventions.
- Write comments to explain complex or non-obvious parts of the code.

#### 8. Security
- Proactively address edge cases, race conditions, and security considerations without being prompted.
- Run `govulncheck` to scan your code for known vulnerabilities before submitting.

#### 9. Performance Considerations
- Highlight potential performance implications and optimization opportunities in suggested code.

#### 10. Architectural Context
- Frame solutions within broader architectural contexts and suggest design alternatives when appropriate.

#### 11. Debugging
- When debugging, provide targeted diagnostic approaches rather than shotgun solutions.

#### 12. Git and Commit Conventions
- **Branching:** Use feature branches with descriptive names (e.g., `feat/new-feature`, `fix/bug-fix`) following `{{branch_naming_convention}}`.
- **Commits:**
    - Keep commits focused on single logical changes to facilitate code review and bisection.
    - Use interactive rebase to clean up history before merging feature branches.
    - Write meaningful commit messages that explain *why* changes were made, not just *what*.
- **Conventional Commits:**
    - Follow the format: `type(scope): description` for all commit messages.
    - Use consistent types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`.
    - Define clear scopes based on `{{project_modules}}` to indicate affected areas.
    - Include issue references in commit messages to link changes to requirements.
    - Use a breaking change footer (`!` or `BREAKING CHANGE:`) to clearly mark incompatible changes.
- **Automation:**
    - Leverage git hooks to enforce code quality checks before commits and pushes.
    - Configure commitlint to automatically enforce the conventional commit format.

By following these guidelines, you will contribute to a high-quality and maintainable codebase.
