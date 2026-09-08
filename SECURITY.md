# Security Policy

## Supported versions

chesssy is developed on `main`, and fixes land there first. Only the latest
release and `main` receive security fixes.

| Version | Supported |
| --- | --- |
| `main` | yes |
| latest release | yes |
| older releases | no |

## Reporting a vulnerability

Report privately, not in a public issue.

Use GitHub's [private vulnerability
reporting](https://github.com/kaitoz11/chesssy/security/advisories/new) on this
repository. That opens a draft advisory visible only to you and the maintainer.

Please include:

- what the problem is and what an attacker gains from it,
- the input that triggers it — a FEN, a UCI command sequence, or a file, exactly
  as it must be fed to the engine,
- the version (`chesssy version`), OS and Go version,
- a crash trace or reproducer if you have one.

Expect an acknowledgement within seven days. A fix, or an explanation of why the
report is not treated as a vulnerability, follows within thirty days for anything
confirmed. Please give the maintainer a chance to publish a fix before disclosing
publicly; credit is given in the advisory and the changelog unless you prefer
otherwise.

## Scope

chesssy is a chess engine: it reads UCI commands from standard input, writes to
standard output, and opens no sockets, no files it was not asked to open, and no
network connections. The realistic attack surface is therefore the parsing of
untrusted input, which matters because a GUI or a tournament manager may feed the
engine positions from an untrusted game file.

In scope:

- memory unsafety or a panic reachable from parsing a FEN, a UCI command line or
  a move list — a panic is a denial of service for whatever is driving the engine,
- unbounded memory or CPU consumption caused by a crafted input, beyond what the
  `Hash` and `go` limits ask for,
- an integer overflow or index-out-of-range in the search or the tables that a
  crafted position can reach,
- anything that makes the engine write outside the paths given on the command
  line.

Out of scope:

- the engine playing a weak move, losing a game, or misjudging a position,
- resource use that follows directly from the options given: a large `Hash` value
  allocating that much memory, or `go infinite` searching until told to stop,
- panics that require modifying the source or calling internal functions with
  arguments the engine itself never produces,
- results from automated scanners with no demonstrated impact on this codebase.

Because chesssy has no dependencies outside the Go standard library, there is no
third-party dependency surface to report against. Vulnerabilities in the Go
toolchain or standard library should go to the [Go security
team](https://go.dev/security/policy); tell us too if a released chesssy binary
needs rebuilding because of one.

## Build and release integrity

The other way a project like this gets compromised is through its build, so the
release path is deliberately small:

- **No dependencies.** The engine imports only the Go standard library, so there
  is no package for an attacker to take over. `go.mod` has no `require` block and
  there is no `go.sum`. A pull request that adds a dependency is a change to this
  threat model and will be treated as one.
- **Actions pinned by digest.** Every GitHub Action is pinned to a full commit
  SHA, with the version in a trailing comment, so moving a tag upstream — the way
  the `tj-actions/changed-files` compromise reached its victims — cannot change
  what runs here. Dependabot proposes updates to those SHAs monthly.
- **No privileged fork triggers.** There is no `pull_request_target` or
  `workflow_run` workflow, so no code from a fork's pull request runs with this
  repository's token or secrets. Pull request workflows get a read-only token.
- **Least privilege.** Workflows default to `contents: read`. Only the release job
  is granted `contents: write`, it runs solely on a tag push or a manual dispatch,
  and it uses two first-party actions and nothing else. `persist-credentials` is
  off for every checkout, so no git credential is left in the workspace.
- **Tools verified by checksum.** staticcheck, govulncheck and GoReleaser are
  fetched with `go run tool@version` at pinned versions, which the module proxy and
  the Go checksum database verify by content hash. A moved upstream tag is inert.
- **Reproducible artefacts.** Release binaries are built with `-trimpath`, with
  `CGO_ENABLED=0` and with the commit timestamp as the module timestamp, and are
  published with a `checksums.txt` you can verify a download against.

If you find a way around any of this — a workflow that can be made to run
attacker-controlled code, or a path to the release token — report it as a
vulnerability through the process above.
