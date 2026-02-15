## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] New plugin (database or storage provider)
- [ ] Performance improvement
- [ ] Code refactoring

## Related Issues

<!-- Link to related issues, e.g., "Fixes #123" or "Relates to #456" -->

Fixes #

## Changes Made

<!-- List the main changes made in this PR -->

-
-
-

## Testing

<!-- Describe the tests you ran to verify your changes -->

- [ ] All existing tests pass (`go test ./...`)
- [ ] Added new unit tests for new functionality
- [ ] Added/updated integration tests
- [ ] Tested manually with the following configuration:

  ```text
  DB_TYPE=...
  STORAGE_TYPE=...
  ```

### Test Results

```text
# Paste relevant test output here
```

## Documentation

<!-- Check all that apply -->

- [ ] Updated README.md
- [ ] Updated EXTENDING.md
- [ ] Added/updated godoc comments
- [ ] Added example usage
- [ ] Updated relevant documentation files

## Checklist

<!-- Ensure all items are checked before submitting -->

- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings or errors
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## For Plugin Contributions

<!-- Only fill out if adding a new plugin -->

- [ ] Implemented required interface (`DatabaseProvider` or `StorageProvider`)
- [ ] Added auto-registration in `init()` function
- [ ] Included unit tests
- [ ] Added integration tests (Docker Compose service if applicable)
- [ ] Updated plugin list in README.md
- [ ] Added example in `examples/` directory
- [ ] Created or updated template in `_templates/` if needed

## Screenshots (if applicable)

<!-- Add screenshots to help explain your changes -->

## Additional Context

<!-- Add any other context about the PR here -->

## Breaking Changes

<!-- If this PR includes breaking changes, describe them here and provide migration guide -->
