## Add or fixing a feature

To add or fix any feature, open a pull request with the changes and wait for the review.

## Testing

To run the tests, run the following command:

```bash
make test
```

## Releasing

To run the release, follow the following steps:

1. Generate a tag for the release

```bash
VERSION=1.0.10 make release
```

2. Draft a new release on the github using the new version

Link: https://github.com/lucasvmiguel/integration/releases/new
