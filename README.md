# Components CLI

## GitHub Copilot Coding Agent Implementation Specification

### Objective

Implement a Go-based JSON component registry and offline updater with extensible adapter support and plugin architecture.

The system manages offline copies of external components. Each component is defined by a JSON manifest. A root registry JSON file references component manifests and defines reusable source and forge configurations.

The implementation must be generic, extensible through adapters and plugins, and support language-specific package registries. Do not hardcode component names, directories, repository paths, or an assumed `lib/` directory. Component manifest paths must come from the root registry configuration.

The system should support source identifiers such as:

```text
github://owner/repository
forgejo://instance-alias/owner/repository
gitea://instance-alias/owner/repository
gitlab://instance-alias/owner/repository
npm://package-name
unpkg://package-name
pypi://package-name
cargo://crate-name
nuget://package-name
maven://group-id:artifact-id
composer://vendor/package
gem://gem-name
pub://package-name
crates-io://crate-name
custom://adapter-name:target
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

The root registry may also define custom adapters:

```json
{
  "adapters": {
    "custom-adapter": {
      "type": "plugin",
      "source": "./plugins/custom-adapter.so"
    }
  },
  "components": {
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
cargo://
nuget://
maven://
composer://
gem://
pub://
```

The first implementation may focus on source parsing and Git-based sources, but unsupported adapters must return clear errors instead of silently treating them as Git repositories.

Optionally support direct archive URLs:

```text
url://https://example.com/archive.tar.gz
```

Keep direct URL behavior separate from forge behavior.

### 6. Adapter architecture

The system must support three types of adapters:

#### 6.1 Built-in adapters

Implement as Go code directly in the binary.

Forge adapters:

- GitHub
- GitLab
- Forgejo
- Gitea

Package registry adapters:

- npm / unpkg (JavaScript/Node.js)
- PyPI (Python)
- Cargo / crates.io (Rust)
- NuGet (.NET)
- Maven (Java)
- Composer (PHP)
- RubyGems (Ruby)
- pub.dev (Dart)

#### 6.2 Plugin adapters

Load external adapters from compiled shared object files (`.so` on Unix, `.dll` on Windows).

Define a plugin interface:

```go
type PluginAdapter interface {
    Name() string
    Version() string
    Type() string
    Capabilities() []string
    Initialize(config map[string]interface{}) error
    Resolve(ctx context.Context, source ResolvedSource, version string) (ResolvedVersion, error)
    ListVersions(ctx context.Context, source ResolvedSource) ([]Version, error)
}
```

Plugins must implement a well-defined export function:

```c
// Expected C export that Go plugins must provide
extern PluginAdapter* NewAdapter(void);
```

Plugins can be registered in the root registry configuration:

```json
{
  "adapters": {
    "my-private-registry": {
      "type": "plugin",
      "source": "./plugins/my-private-registry.so",
      "config": {
        "url": "https://private-registry.example.com",
        "auth": {
          "secret": "private-registry"
        }
      }
    }
  }
}
```

#### 6.3 HTTP-based adapters

Support remote adapters that expose an HTTP API.

Define an HTTP adapter contract:

```json
{
  "type": "http",
  "url": "http://localhost:9999",
  "operations": {
    "list-versions": "GET /versions/{namespace}/{name}",
    "resolve-version": "GET /resolve/{namespace}/{name}/{version}",
    "download": "GET /download/{namespace}/{name}/{version}"
  }
}
```

Register in root config:

```json
{
  "adapters": {
    "remote-registry": {
      "type": "http",
      "url": "https://adapter-service.example.com",
      "timeout": "30s",
      "auth": {
        "secret": "remote-registry"
      }
    }
  }
}
```

### 7. Common adapter interface

Create a unified interface for all adapters:

```go
type SourceAdapter interface {
    Type() string
    Name() string
    
    RepositoryURL(
        config SourceConfig,
        repository string,
    ) string
    
    CloneURL(
        config SourceConfig,
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
    
    DownloadArtifact(
        ctx context.Context,
        source ResolvedSource,
        version ResolvedVersion,
        destination string,
    ) error
    
    Capabilities() []string
}
```

The interface must support:

- Version discovery.
- Version resolution.
- Artifact download.
- Metadata retrieval.
- Authentication.
- Custom configuration.

### 8. Language-specific package registry adapters

Each major language registry must be supported through a dedicated adapter.

#### 8.1 npm / unpkg (JavaScript/Node.js)

```bash
npm://lodash
npm://lodash@4.17.21
unpkg://react
```

- Resolve from npm registry (https://registry.npmjs.org).
- Support version ranges.
- Download tarballs from npm CDN.

#### 8.2 PyPI (Python)

```bash
pypi://requests
pypi://django@3.2.0
```

- Query PyPI JSON API.
- Support semantic versioning.
- Download wheels or source distributions.

#### 8.3 Cargo / crates.io (Rust)

```bash
cargo://serde
cargo://tokio@1.0.0
crates-io://regex
```

- Query crates.io API.
- Download crate archives.

#### 8.4 NuGet (.NET)

```bash
nuget://Newtonsoft.Json
nuget://EntityFramework@6.4.4
```

- Query nuget.org API.
- Download packages.

#### 8.5 Maven (Java)

```bash
maven://org.slf4j:slf4j-api
maven://com.google.guava:guava@30.0
```

- Query Maven Central Repository.
- Support version resolution.
- Download JARs.

#### 8.6 Composer (PHP)

```bash
composer://laravel/framework
composer://symfony/console@5.0.0
```

- Query Packagist API.
- Download packages.

#### 8.7 RubyGems (Ruby)

```bash
gem://rails
gem://sinatra@2.1.0
```

- Query rubygems.org API.
- Download gems.

#### 8.8 pub.dev (Dart)

```bash
pub://provider
pub://flutter@latest
```

- Query pub.dev API.
- Download packages.

Each adapter must:

- Implement the unified `SourceAdapter` interface.
- Support version discovery and resolution.
- Resolve package names to downloadable artifacts.
- Record the source package name, version, and download URL in the lock file.
- Validate downloaded artifacts when checksums are available.

### 9. Plugin development guide

To be included in documentation:

- Example plugin structure and build process.
- Required exports and function signatures.
- Configuration schema for custom adapters.
- Error handling and logging conventions.
- Testing harness for plugins.
- Example implementations for private registries.

### 10. Adapter API and discovery

The system must provide a queryable adapter registry:

```bash
components adapter list
components adapter show <name>
components adapter capabilities <name>
```

Output:

```text
NAME                    TYPE        CAPABILITIES
github                  builtin     list-versions, resolve, clone, auth
npm                     builtin     list-versions, resolve, download, auth
pypi                    builtin     list-versions, resolve, download
my-plugin               plugin      list-versions, resolve, download
remote-registry         http        list-versions, resolve, download
```

The implementation must:

- Enumerate all loaded adapters.
- Report adapter capabilities.
- Show configuration requirements.
- Validate adapter configuration.
- Report adapter-specific errors clearly.

### 11. Version resolution

Support these version forms:

```text
1.2.3
v1.2.3
^1.2.3
~1.2.3
>=1.2.0 <2.0.0
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

### 12. Offline storage

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

### 13. Lock file

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
      "adapter": "github",
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

### 14. Download and installation lifecycle

Use this lifecycle:

```text
Load root registry
    ↓
Load adapters (built-in, plugins, HTTP)
    ↓
Load component manifest
    ↓
Parse source URI
    ↓
Resolve source adapter
    ↓
Resolve requested version using adapter
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

### 15. Verification and validation

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

### 16. CLI

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
components adapter list
components adapter show <name>
components adapter capabilities <name>
```

Required options:

```bash
--file <path>
--storage <path>
--lock <path>
--secrets <path>
--plugins-dir <path>
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
4. Create an empty `adapters` object.
5. Create a secrets file or secrets directory using the configured secrets path.
6. Add the secrets path to `.gitignore`.
7. Create a default lock file if requested or required by the selected configuration.
8. Refuse to overwrite existing files unless an explicit `--force` option is supplied.
9. Use relative paths where possible so the registry remains portable.

Example:

```bash
components init
components init --file config/components.json --storage .cache/components --plugins-dir ./plugins
```

#### `add` command

`components add` must make adding a component easy without requiring manual JSON editing.

Example:

```bash
components add Ruffle github://ruffle-rs/ruffle
components add Ionicons npm://ionicons --version 7.1.0
components add InternalTools company://platform/internal-tools --path vendor/internal-tools
components add Django pypi://django --version 4.0.0
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
3. Load the existing root registry and adapters.
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

### 17. Secrets and authentication

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
  "pypi": {
    "username": "myuser",
    "password": "secret"
  },
  "npm": {
    "token": "npm-token"
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
components secret set pypi.username
components secret set npm.token
components secret list
components secret remove company.token
```

`secret set` should prompt interactively when no value is supplied, and should avoid accepting secrets through shell arguments by default. A non-interactive option may read from standard input or an environment variable.

`secret list` must show only names, never values:

```text
github.token
pypi.username
npm.token
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

Package registry adapters should support configured token authentication. Direct Git sources should support the selected secure Git credential mechanism without writing credentials into clone URLs.

### 18. Go package layout

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
├── adapter/
│   ├── interface.go
│   ├── registry.go
│   ├── loader.go
│   ├── builtin/
│   │   ├── github.go
│   │   ├── gitlab.go
│   │   ├── forgejo.go
│   │   ├── gitea.go
│   │   ├── npm.go
│   │   ├── pypi.go
│   │   ├── cargo.go
│   │   ├── nuget.go
│   │   ├── maven.go
│   │   ├── composer.go
│   │   ├── gem.go
│   │   └── pub.go
│   ├── plugin/
│   │   ├── loader.go
│   │   ├── interface.go
│   │   └── sandbox.go
│   └── http/
│       └── remote.go
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
    ├── adapter.go
    ├── secret.go
    └── update.go

pkg/
└── adapter/
    └── api.go
```

The agent may adjust the package layout if the repository has an established Go structure.

### 19. Configuration validation

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
- Missing plugin files.
- Invalid plugin configuration.
- Unavailable HTTP adapters.

Errors must identify the relevant file and JSON field. For example:

```text
component "Example": both "ver" and "version" are present with different values
adapter "my-plugin": plugin file not found at ./plugins/my-plugin.so
```

### 20. Security requirements

- Never store credentials in JSON manifests.
- Read authentication tokens from the local secrets store or environment variables.
- Redact credentials from logs and errors.
- Use HTTPS by default for HTTP sources and HTTP adapters.
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
- Sandbox plugin execution to prevent filesystem escape.
- Validate plugin signatures when available.
- Apply resource limits to plugin operations (memory, time, network).

### 21. Adapter extension point

The system must provide a public API for custom adapter development:

```bash
go get github.com/Evoker-Industries/components-cli/pkg/adapter
```

Example custom adapter skeleton:

```go
package main

import (
    "github.com/Evoker-Industries/components-cli/pkg/adapter"
)

type MyAdapter struct {
    config map[string]interface{}
}

func (a *MyAdapter) Name() string { return "my-registry" }
func (a *MyAdapter) Type() string { return "plugin" }
func (a *MyAdapter) Capabilities() []string {
    return []string{"list-versions", "resolve", "download"}
}

func (a *MyAdapter) Initialize(cfg map[string]interface{}) error {
    a.config = cfg
    return nil
}

// Implement remaining methods...

var AdapterInstance = &MyAdapter{}
```

Build as a plugin:

```bash
go build -o my-registry.so -buildmode=plugin .
```

### 22. Testing requirements

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
- Adapter registration and discovery.
- Adapter capability reporting.
- Plugin loading and initialization.
- HTTP adapter negotiation.
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
- Package registry API responses (npm, PyPI, Cargo, etc.).
- Release discovery.
- Artifact downloads.
- Failed downloads.
- Authentication headers.
- Retry behavior.
- Offline mode.
- Secret lookup by forge alias.
- Plugin lifecycle.
- HTTP adapter communication.

Add end-to-end tests using temporary directories and temporary registry files. Test manifest paths in multiple arbitrary locations to ensure the implementation does not depend on `lib/`. Test initialization, adding components with various adapters, and loading plugins.

### 23. Built-in adapter test requirements

Each built-in adapter must have comprehensive tests including:

- Version discovery.
- Version resolution.
- Artifact download.
- Error handling for network failures.
- Authentication handling.
- Metadata extraction.
- Cache-friendly requests.

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
10. npm, PyPI, Cargo, NuGet, Maven, Composer, RubyGems, and pub.dev adapters are implemented.
11. Direct Git URLs are supported.
12. Components can be downloaded or cloned into configurable storage.
13. Installed versions are immutable.
14. A lock file records resolved versions, checksums, and adapter information.
15. Failed updates do not replace the current version.
16. Offline commands work for already-installed components.
17. Unsafe archive paths are rejected.
18. `components init` creates a usable new registry without overwriting files by default.
19. `components add` creates a component manifest and registers it using a relative `ref`.
20. `components add` works with arbitrary manifest paths and does not assume `lib/`.
21. Secrets can be added, listed by name, and removed through CLI commands.
22. Secrets are stored outside manifests and lock files in a locally ignored file.
23. Secrets files use restrictive permissions and are atomically updated.
24. All major language package registries support authenticated requests through secrets.
25. Secrets never appear in logs, command output, generated files, or repository URLs.
26. Plugin adapters can be loaded from `.so` files and configured in the root registry.
27. HTTP adapters can be discovered and used for remote adapter implementations.
28. `components adapter list`, `components adapter show`, and `components adapter capabilities` work.
29. Custom adapters can be developed using the public API and packaged as plugins.
30. Adapters are registered and resolved dynamically based on source URI scheme.
31. The lock file includes the adapter name for each installed component.
32. Tests cover arbitrary manifest locations, adapters, and plugins.
33. `go test ./...` passes.
34. The CLI builds successfully with:

```bash
go build ./cmd/components
```

35. The public adapter API is available via:

```bash
go get github.com/Evoker-Industries/components-cli/pkg/adapter
```

36. Documentation includes example custom adapter implementations.
