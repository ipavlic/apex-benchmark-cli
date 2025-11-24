# Contributing

## Setup

**Requirements:** Go 1.21+, Salesforce CLI (`sf`)

```bash
git clone https://github.com/ipavlic/apex-benchmark-cli.git
cd apex-benchmark-cli
go mod download
make build
make test
```

## Contributing

**Bug Reports:**
- Check existing issues first
- Include: description, steps to reproduce, environment details

**Pull Requests:**
1. Fork and create feature branch
2. Write tests, ensure they pass (`make test`)
3. Format code (`go fmt`)
4. Commit with clear messages
5. Push and create PR

## Guidelines

**Code:**
- Follow Go conventions
- Add tests for new features (aim for >80% coverage)
- Comment exported functions
- Use table-driven tests

**Commits:**
```
Add feature: brief description

- Detail 1
- Detail 2
```

**Architecture:**
See [DESIGN.md](DESIGN.md) for details. Key principles:
- Separation of concerns
- Testability
- Extensibility

## First Contribution?

Look for `good-first-issue` tags:
- Documentation improvements
- Small bug fixes
- Test coverage
- Example snippets

## Questions?

Open a [Discussion](https://github.com/ipavlic/apex-benchmark-cli/discussions) or comment on an issue.

---

Thanks for contributing!
