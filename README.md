# Components CLI

## GitHub Copilot Coding Agent Implementation Specification

### Objective

Implement a Go-based JSON component registry and offline updater.

The system manages offline copies of external components. Each component is defined by a JSON manifest. A root registry JSON file references component manifests and defines reusable source and forge configurations.

The implementation must be generic. Do not hardcode component names, directories, repository paths, or an assumed `lib/` directory. Component manifest paths must come from the root registry configuration.

The system should support source identifiers such as:

```text
github://owner/repository
forgejo://instance-alias/owner/repository
gitea://instance-alias/owner/repository
gitlab://instance-alias/owner/repository
npm://package-name
unpkg://package-name
git+https://git.example.com/owner/repository.git
```

---

## Existing configuration format

The root registry currently uses this general structure:

```json
{
  "Components": {
    "ComponentName": {
      "ref": "./path/to/component.json"
    }
  }
}
```

A referenced component manifest has this structure:

```json
{
  "component": {
    "name": "Component Name",
    "source": "github://owner/repository",
    "ver": "1.0.0",
    "dir": "."
  }
}
```

The implementation must support the existing abbreviated fields:

- `ver`
- `dir`

It should also support these preferred aliases:

- `version`
- `directory`

If both aliases are present with different values, validation must fail.

The root property may be named `Components` for compatibility. New generated configuration should use lowercase `components`.

---

## Functional requirements

### 1. Root registry loading

Load a configurable root registry file. The default may be `components.json`, but the path must be configurable through a CLI option such as:

```bash
components --file path/to/registry.json validate
```

Do not assume that component manifests are located under `lib/`.

Support both:

```json
{
  "components": {
    "Example": {
      "ref": "./some/path/component.json"
    }
  }
}
```

and:

```json
{
  "Components": {
    "Example": {
      "ref": "./some/path/component.json"
    }
  }
}
```

Resolve each `ref` relative to the directory containing the root registry file, not relative to the process working directory.

### 2. Forge configuration

The root registry may define reusable forge instances:

```json
{
  "forges": {
    "github": {
      "type": "github",
      "url": "https://github.com",
      "api": "https://api.github.com"
    },
    "codeberg": {
      "type": "forgejo",
      "url": "https://codeberg.org",
      "api": "https://codeberg.org/api/v1"
    },
    "company": {
      "type": "forgejo",
      "url": "https://git.company.example",
      "api": "https://git.company.example/api/v1"
    }
  },
  "components": {
    "Example": {
      "ref": "./components/example/component.json"
    }
  }
}
```

Forge configuration must be generic and reusable. Do not hardcode public or private forge hosts.

Recommended model:

```go
type ForgeConfig struct {
    Type string `json:"type"`
    URL  string `json:"url"`
    API  string `json:"api,omitempty"`
}
```

### 3. Component manifest loading

Load every manifest using its root-level `ref`. Component paths may be arbitrary, for example:

```text
components/example.json
vendor/ruffle/component.json
third_party/icons/ionicons.json
packages/runtime/manifest.json
```

The loader must not infer a component path from its name.

Recommended model:

```go
type Component struct {
    Name       string        `json:"name"`
    Source     string        `json:"source"`
    Version    string        `json:"version,omitempty"`
    Ver        string        `json:"ver,omitempty"`
    Directory  string        `json:"directory,omitempty"`
    Dir        string        `json:"dir,omitempty"`
    Update     *UpdatePolicy `json:"update,omitempty"`
    Verify     *VerifyPolicy `json:"verify,omitempty"`
    Validation *Validation   `json:"validation,omitempty"`
}
```

Normalize `ver` and `version` into one internal `Version` field. Normalize `dir` and `directory` into one internal `Directory` field.

Required fields:

- `name`
- `source`

`version` may be optional for sources that use branches, rolling channels, or latest-version resolution.

### 4. Source URI parsing

Implement a generic source URI parser.

Input:

```text
company://platform/internal-tools
```

Output:

```text
scheme: company
target: platform/internal-tools
```

Input:

```text
git+https://git.example.com/team/project.git
```

Output:

```text
scheme: git+https
target: git.example.com/team/project.git
```

The parser must:

- Require a valid scheme.
- Require `://`.
- Normalize the scheme to lowercase.
- Preserve the target.
- Reject empty targets.
- Reject control characters.
- Reject malformed source strings.
- Avoid unsafe path normalization.

Recommended type:

```go
type SourceURI struct {
    Scheme string
    Target string
}
```

Do not implement source parsing with a hardcoded list of component names or repositories.

### 5. Source resolution

Resolve sources based on the parsed scheme.

For:

```text
github://owner/repository
```

resolve `github` from the root `forges` configuration.

For:

```text
company://team/project
```

resolve `company` from the root `forges` configuration.

The alias before `://` must first be checked against configured forge aliases.

Recommended resolved model:

```go
type ResolvedSource struct {
    Type       string
    Alias      string
    BaseURL    string
    APIURL     string
    Repository string
    CloneURL   string
}
```

Support direct Git sources without a forge entry:

```text
git+https://git.example.com/owner/repository.git
git+ssh://git@git.example.com/owner/repository.git
```

Create adapter interfaces for package sources:

```text
npm://
unpkg://
pypi://
```

The first implementation may focus on source parsing and Git-based sources, but unsupported adapters must return clear errors instead of silently treating them as Git repositories.

Optionally support direct archive URLs:

```text
url://https://example.com/archive.tar.gz
```

Keep direct URL behavior separate from forge behavior.

### 6. Forge adapters

Create a common interface for forge providers:

```go
type ForgeAdapter interface {
    Type() string

    RepositoryURL(
        config ForgeConfig,
        repository string,
    ) string

    CloneURL(
        config ForgeConfig,
        repository string,
    ) string

    ListVersions(
        ctx context.Context,
        source ResolvedSource,
    ) ([]Version, error)

    ResolveVersion(
        ctx context.Context,
        source ResolvedSource,
        requested string,
    ) (ResolvedVersion, error)
}
```

Implement adapters for:

- GitHub
- GitLab
- Forgejo
- Gitea

Forgejo and Gitea may share implementation details, but register them as separate provider types. Custom hosts must work through root configuration.

### 7. Version resolution

Support these version forms:

```text
1.2.3
v1.2.3
latest
stable
nightly
main
branch:main
tag:v1.2.3
commit:abcdef123
```

Do not assume every source uses semantic versioning.

Store both the requested version and the resolved version or commit. For Git sources, record the final commit SHA. For release-based sources, record the release ID when available, release tag, published timestamp, and selected asset.

### 8. Offline storage

Store each installed version in an immutable version-specific directory. The storage root must be configurable.

Example layout:

```text
<storage-root>/
└── component-name/
    ├── 1.0.0/
    ├── 1.1.0/
    └── current
```

Safely normalize component names before using them as filesystem paths. Never overwrite an existing resolved version in place.

The `current` marker may be implemented as a symlink, a text file containing the active version, or an internal metadata record. Hide platform-specific behavior behind the storage implementation.

### 9. Lock file

Maintain a lock file containing exact resolved state. Its path must be configurable; `components.lock.json` is a suitable default.

Example:

```json
{
  "schema": "component-lock/v1",
  "generatedAt": "2026-10-01T12:00:00Z",
  "components": {
    "Example": {
      "source": "github://owner/repository",
      "requestedVersion": "0.6.0",
      "resolvedVersion": "0.6.0",
      "commit": "abcdef1234567890",
      "path": "stored/example/0.6.0",
      "verified": true,
      "files": [
        {
          "path": "example.wasm",
          "sha256": "..."
        }
      ]
    }
  }
}
```

Write the lock file atomically. Do not update it until download, verification, and validation succeed.

### 10. Download and installation lifecycle

Use this lifecycle:

```text
Load root registry
    ↓
Load component manifest
    ↓
Parse source URI
    ↓
Resolve source adapter
    ↓
Resolve requested version
    ↓
Download or clone into staging directory
    ↓
Verify checksums and signatures
    ↓
Validate extracted files
    ↓
Create manifest
    ↓
Move staging directory into immutable storage
    ↓
Update lock file atomically
    ↓
Update current pointer
```

If any step fails:

- Preserve the current installed version.
- Do not corrupt the lock file.
- Remove or retain staging data according to a cleanup option.
- Return a non-zero exit status.

### 11. Verification and validation

Support SHA-256 verification. At minimum, record hashes for installed files in the generated manifest.

Reject unsafe archive entries such as:

```text
../../etc/passwd
/absolute/path
C:\Windows\System32\file
```

All extracted paths must remain within the staging directory.

Support optional validation such as:

```json
{
  "component": {
    "name": "Example",
    "source": "github://owner/repository",
    "version": "1.0.0",
    "verify": {
      "sha256": true
    },
    "validation": {
      "requiredFiles": [
        "dist/index.js",
        "dist/index.css"
      ]
    }
  }
}
```

The validation system must not assume any particular component or filename.

### 12. CLI

Implement a Go CLI named `components`.

Required commands:

```bash
components init
components add <name> <source>
components validate
components list
components show <name>
components check
components install <name>
components update
components update <name>
components current <name>
components rollback <name>
components clean
components secret set <name> <value>
components secret list
components secret remove <name>
```

Required options:

```bash
--file <path>
--storage <path>
--lock <path>
--secrets <path>
--offline
--verbose
```

The CLI must use configured paths and must not assume repository-specific directories.

#### `init` command

`components init` must initialize a new component registry in the selected directory.

Expected behavior:

1. Create the root registry file if it does not exist.
2. Create an empty `components` object.
3. Create an empty `forges` object.
4. Create a secrets file or secrets directory using the configured secrets path.
5. Add the secrets path to `.gitignore`.
6. Create a default lock file if requested or required by the selected configuration.
7. Refuse to overwrite existing files unless an explicit `--force` option is supplied.
8. Use relative paths where possible so the registry remains portable.

Example:

```bash
components init
components init --file config/components.json --storage .cache/components
```

#### `add` command

`components add` must make adding a component easy without requiring manual JSON editing.

Example:

```bash
components add Ruffle github://ruffle-rs/ruffle
components add Ionicons unpkg://ionic-team/ionicons --version 7.1.0
components add InternalTools company://platform/internal-tools --path vendor/internal-tools
```

Supported options should include:

```bash
--version <version>
--path <manifest-path>
--directory <storage-directory>
--forge <alias>
--yes
```

The command must:

1. Validate the component name.
2. Validate and parse the source URI.
3. Load the existing root registry.
4. Reject duplicate component names unless `--force` is provided.
5. Create a component manifest at the requested or generated path.
6. Add a root registry entry with a relative `ref`.
7. Preserve existing formatting and entries as much as practical.
8. Write registry and manifest files atomically.
9. Optionally install the component when `--install` is supplied.
10. Never place credentials in the generated component manifest.

The command must not derive paths from a hardcoded `lib/` convention. If no manifest path is supplied, use a configurable default template such as `<components-directory>/<safe-name>/component.json`.

Example generated manifest:

```json
{
  "component": {
    "name": "Ruffle",
    "source": "github://ruffle-rs/ruffle",
    "version": "latest",
    "directory": "."
  }
}
```

Example generated root entry:

```json
{
  "components": {
    "Ruffle": {
      "ref": "./components/ruffle/component.json"
    }
  }
}
```

### 13. Secrets and authentication

Provide a secure, local secrets feature for credentials required by GitHub, Forgejo, Gitea, GitLab, package registries, or private Git repositories.

Secrets must not be stored in component manifests, the root registry, or the lock file.

The secrets path must be configurable through `--secrets` and may default to a file such as:

```text
.components.secrets.json
```

The default secrets path must always be added to `.gitignore`. If the user selects a custom secrets path, that path must also be added to `.gitignore` unless the user explicitly disables this behavior.

Recommended secrets format:

```json
{
  "github": {
    "token": "ghp_example"
  },
  "company": {
    "token": "forgejo-token"
  },
  "private-git": {
    "username": "git-user",
    "password": "secret"
  }
}
```

The implementation must:

- Create the secrets file with restrictive permissions, such as `0600` on Unix-like systems.
- Refuse to use a secrets file that is group/world writable unless explicitly overridden.
- Never print secret values.
- Redact tokens and passwords from logs and error messages.
- Use secrets only at runtime to construct authenticated requests or Git credentials.
- Keep secrets out of generated manifests and lock files.
- Support environment variables as an optional higher-precedence source.
- Make it possible to use different credentials for different forge aliases.
- Preserve unknown secret fields when updating the file.
- Write the secrets file atomically.

Recommended commands:

```bash
components secret set github.token
components secret set company.token
components secret set private-git.username
components secret set private-git.password
components secret list
components secret remove company.token
```

`secret set` should prompt interactively when no value is supplied, and should avoid accepting secrets through shell arguments by default. A non-interactive option may read from standard input or an environment variable.

`secret list` must show only names, never values:

```text
github.token
company.token
private-git.username
private-git.password
```

Forge configuration may reference a secret namespace without embedding credentials:

```json
{
  "forges": {
    "company": {
      "type": "forgejo",
      "url": "https://git.company.example",
      "api": "https://git.company.example/api/v1",
      "auth": {
        "secret": "company"
      }
    }
  }
}
```

The loader must resolve `auth.secret` against the local secrets store at runtime.

GitHub should support token-based API authentication. Forgejo, Gitea, and GitLab should support configured token authentication. Direct Git sources should support the selected secure Git credential mechanism without writing credentials into clone URLs.

### 14. Go package layout

Use a maintainable package structure similar to:

```text
cmd/
└── components/
    └── main.go

internal/
├── config/
│   ├── loader.go
│   ├── types.go
│   └── validator.go
├── component/
│   ├── loader.go
│   ├── normalize.go
│   └── types.go
├── source/
│   ├── parser.go
│   ├── resolver.go
│   └── types.go
├── forge/
│   ├── adapter.go
│   ├── registry.go
│   ├── github.go
│   ├── gitlab.go
│   ├── forgejo.go
│   └── gitea.go
├── downloader/
│   ├── http.go
│   ├── git.go
│   ├── archive.go
│   └── checksum.go
├── storage/
│   ├── filesystem.go
│   ├── staging.go
│   └── manifest.go
├── lockfile/
│   ├── loader.go
│   └── writer.go
├── secrets/
│   ├── store.go
│   ├── file.go
│   ├── permissions.go
│   └── redact.go
├── initcmd/
│   └── init.go
├── addcmd/
│   └── add.go
├── validation/
│   ├── config.go
│   ├── files.go
│   └── archive.go
└── cli/
    ├── init.go
    ├── add.go
    ├── validate.go
    ├── list.go
    ├── check.go
    ├── install.go
    ├── secret.go
    └── update.go
```

The agent may adjust the package layout if the repository has an established Go structure.

### 15. Configuration validation

Validation must detect:

- Invalid JSON.
- Missing component references.
- Missing component manifest files.
- Duplicate component names.
- Missing component names.
- Missing sources.
- Conflicting `ver` and `version`.
- Conflicting `dir` and `directory`.
- Unknown forge aliases.
- Unsupported forge types.
- Invalid forge URLs.
- Unsafe paths.
- Unsupported source schemes.
- Duplicate forge aliases.
- Invalid secret references.
- Secrets paths that are not ignored by Git.

Errors must identify the relevant file and JSON field. For example:

```text
component "Example": both "ver" and "version" are present with different values
```

### 16. Security requirements

- Never store credentials in JSON manifests.
- Read authentication tokens from the local secrets store or environment variables.
- Redact credentials from logs and errors.
- Use HTTPS by default for HTTP sources.
- Do not execute commands defined by JSON.
- Do not allow arbitrary shell hooks in component manifests.
- Prevent archive path traversal.
- Prevent writes outside configured storage and staging directories.
- Use context cancellation for network and Git operations.
- Apply HTTP timeouts.
- Stream large downloads instead of loading them entirely into memory.
- Create secrets files with restrictive permissions.
- Ensure secrets files are ignored by Git.
- Do not pass credentials in repository URLs or command-line arguments.

### 17. Testing requirements

Add unit tests for:

- Root registry loading.
- Relative reference resolution.
- Legacy `Components` support.
- Arbitrary manifest paths.
- `ver` and `version` normalization.
- `dir` and `directory` normalization.
- Source URI parsing.
- Forge alias resolution.
- Custom forge hosts.
- GitHub URL generation.
- Forgejo URL generation.
- Gitea URL generation.
- GitLab URL generation.
- Unsupported source handling.
- Path traversal protection.
- Atomic lock file writes.
- Version comparison.
- `init` file creation.
- `init` refusal to overwrite without `--force`.
- `add` manifest generation.
- `add` root registry updates.
- Duplicate component handling.
- Relative path generation.
- Secrets serialization.
- Secrets file permissions.
- Secret redaction.
- `.gitignore` updates.

Add integration tests using `httptest.Server` for:

- Forge API responses.
- Release discovery.
- Artifact downloads.
- Failed downloads.
- Authentication headers.
- Retry behavior.
- Offline mode.
- Secret lookup by forge alias.

Add end-to-end tests using temporary directories and temporary registry files. Test manifest paths in multiple arbitrary locations to ensure the implementation does not depend on `lib/`. Test initialization and adding components in an empty directory, an existing registry, and a registry with custom forge aliases.

---

## Acceptance criteria

The implementation is complete when:

1. The root registry file can be located using a CLI option.
2. Component manifest paths are resolved from `ref`.
3. No code assumes components live under `lib/`.
4. Both `Components` and `components` root keys are supported.
5. Both abbreviated and full component field names are supported.
6. Forge instances are defined in the root registry.
7. Custom Forgejo, Gitea, and GitLab hosts work through configuration.
8. Source strings are parsed through a generic URI parser.
9. GitHub, GitLab, Forgejo, and Gitea adapters are implemented.
10. Direct Git URLs are supported.
11. Components can be downloaded or cloned into configurable storage.
12. Installed versions are immutable.
13. A lock file records resolved versions and checksums.
14. Failed updates do not replace the current version.
15. Offline commands work for already-installed components.
16. Unsafe archive paths are rejected.
17. `components init` creates a usable new registry without overwriting files by default.
18. `components add` creates a component manifest and registers it using a relative `ref`.
19. `components add` works with arbitrary manifest paths and does not assume `lib/`.
20. Secrets can be added, listed by name, and removed through CLI commands.
21. Secrets are stored outside manifests and lock files in a locally ignored file.
22. Secrets files use restrictive permissions and are atomically updated.
23. GitHub, Forgejo, Gitea, GitLab, and private Git authentication can use configured secrets.
24. Secrets never appear in logs, command output, generated files, or repository URLs.
25. Tests cover arbitrary manifest locations and do not rely on example component names.
26. `go test ./...` passes.
27. The CLI builds successfully with:

```bash
go build ./cmd/components
```
