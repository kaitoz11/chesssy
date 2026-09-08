## What this changes

<!-- One paragraph. What the change does and why. Link the issue if there is one. -->

Closes #

## Measurements

<!--
Required for anything touching search, evaluation or move generation. Delete this
section for documentation, build and tooling changes.

State the machine, and give before/after for whichever numbers the change can move:
node count (make suite), agreement (make agreement), wall clock (make measure).
A change that cuts nodes while agreement falls is pruning too much.
-->

Machine:

| Measurement | Before | After |
| --- | --- | --- |
| `make suite` nodes | | |
| `make agreement` | | |
| `make measure` bench depth 13 | | |

## Checklist

- [ ] `make check` passes (format, vet, staticcheck, tests)
- [ ] `make test-race` passes, if the change touches shared state or threads
- [ ] `make perft` passes, if the change touches move generation
- [ ] `make perft-deep` passes, if node counts could have changed
- [ ] Tests added or updated for the behaviour that changed
- [ ] `CHANGELOG.md` updated under Unreleased, for anything user visible
- [ ] File header comments and `doc.go` reading order updated, for new files
- [ ] Any code taken from elsewhere is GPL compatible, and its source is named above
