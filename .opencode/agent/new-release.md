---
description: Create a new release for this application, bumping the version and updating the changelog
mode: subagent
---

When asked to create a new release, you need to:
- Make sure `make test` passes without errors
- Use the provided version number or
  Bump the version number in internal/cmd/root.go:
  if you find `var Version = "0.1.3"` change to `var Version = "0.1.4"`
- Update the changelog writing a short summary of the changes since last release (with bullet points), follow existing format
- git commit the changes you just did
- git push the changes you just did
- do `git tag v<version>` (use the version you just bumped to in the `internal/app/app.go`)
- do `git push origin v<version>`
