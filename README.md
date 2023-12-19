# <img alt=logo src=etcha.png width=40px> Etcha

> Infinite scale configuration management for distributed applications.

[![Integration](https://github.com/candiddev/etcha/actions/workflows/integration.yaml/badge.svg?branch=main)](https://github.com/candiddev/etcha/actions/workflows/integration.yaml)

[:book: Docs](https://etcha.dev/docs/)\
[:motorway: Roadmap](https://github.com/orgs/candiddev/projects/6/views/37)

Etcha is a source available command line (CLI) tool for building, testing, linting, and distributing application patterns.

Etcha makes distributed applications easy:

- Define your configurations in Jsonnet as Patterns
- Lint and test your patterns
- Build your patterns as release artifacts
- Pull your release artifacts from anywhere or push them on demand
- Use webhook and event handlers to dynamically run commands and Patterns

Visit https://etcha.dev for more information.

## License

The code in this repository is licensed under a personal, non-production [source-available license](./LICENSE.md).  Visit https://etcha.dev/pricing/ for additional licensing options.

## Development

Our development process is mostly trunk-based with a `main` branch that folks can contribute to using pull requests.  We tag releases as necessary using CalVer.

### Repository Layout

- `./github:` Reusable GitHub Actions
- `./go:` Etcha code
- `./hugo:` Etcha website
- `./shell:` Development tooling
- `./shared:` Shared libraries from https://github.com/candiddev/shared

Make sure you initialize the shared submodule:

```bash
git submodule update --init
```

### CI/CD

We use GitHub Actions to lint, test, build, release, and deploy the code.  You can view the pipelines in the `.github/workflows` directory.  You should be able to run most workflows locally and validate your code before opening a pull request.

### Tooling

Visit [shared/README.md](shared/README.md) for more information.
