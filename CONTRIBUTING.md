# Contributing to Imagineer

## Welcome

Thank you for your interest in contributing to Imagineer! We welcome contributions from the community and are excited to have you join us in building a better TTRPG campaign management platform.

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and considerate in all interactions
- Provide constructive feedback
- Welcome newcomers and help them get started
- Focus on what is best for the community

## How to Contribute

### Reporting Bugs

Found a bug? Please open a GitHub issue with:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Your environment (OS, browser, versions)

### Suggesting Features

Have an idea? We'd love to hear it! Open a GitHub issue and describe:

- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Submitting Pull Requests

Ready to contribute code? Follow the process outlined below.

## Development Setup

1. **Fork the repository** to your GitHub account
2. Clone your fork locally
3. Follow the Quick Start in the README to set up your environment
4. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/my-feature
   ```
5. Keep commits focused and atomic - each commit should represent a single logical change

## Pull Request Process

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```
2. Make your changes following our code standards
3. Write or update tests for your changes
4. Run the test suite to ensure everything passes:
   ```bash
   make test-all
   ```
5. Update `CHANGELOG.md` with a description of your changes
6. Submit your PR with a clear description of:
   - What changes you made
   - Why you made them
   - Any relevant context or screenshots
7. CodeRabbit will automatically review your PR

## Code Standards

### Go

- Follow the standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- Run `gofmt` before committing
- Use 4-space indentation

### React/TypeScript

- Enable TypeScript strict mode
- Follow ESLint rules
- Keep components focused and reusable

### General

- Include copyright headers on new files
- Write tests for new functionality
- Search for duplicates before creating entities
- Never hardcode configuration values

## Commit Messages

We use conventional commits style:

```
type(scope): brief description

Longer description if needed.

Co-Authored-By: Name <email>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

When using AI assistance, include the Co-Authored-By line.

## Getting Help

- **Bugs and Features**: Open a GitHub Issue
- **Questions and Discussion**: Use GitHub Discussions

We look forward to your contributions!
