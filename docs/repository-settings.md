# Repository settings

After the workflow in this change exists on the default branch, protect `main`
with a branch ruleset. Repository settings are not versioned by a pull request,
so an administrator must enable these controls in GitHub:

- require pull requests before merging;
- require at least one approval and dismiss stale approvals;
- require a code-owner review;
- require conversation resolution;
- require the `Required / Repository policy` status check;
- block force pushes and branch deletion;
- do not allow bypass for repository administrators.

The existing `Quality / PR tests and local reports` workflow is intentionally
path-filtered and therefore should not be the only required check: documentation
pull requests do not create that check. The always-running repository-policy
job supplies a stable required-check context, while Quality continues to run
tests and reports for source and workflow changes.
