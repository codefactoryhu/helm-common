# Contributing to helm-common

Thanks for your interest in helm-common library chart. Our goal is to bring the common repetitive tasks and solutions of the helm charts into this repository and reuse them.

## Getting Started

An easy way to get started helping the project is to *file an issue*. You can do that on the CF Helm Common issues page by clicking on the *New Issue* button at the right. Issues can include bugs to fix, features to add, or documentation that looks outdated.

## Contributions

CF Helm Common welcomes contributions from everyone.

Contributions to CF Helm Common should be made in the form of GitLab merge requests. Each merge request will
be reviewed by a core contributor (someone with permission to land patches) and either landed in the
main tree or given feedback for changes that would be required.

## Merge Request Checklist

- Branch from the master branch and, if needed, rebase to the current master
  branch before submitting your merge request. If it doesn't merge cleanly with
  master you may be asked to rebase your changes.

- Commits should be as small as possible, while ensuring that each commit is
  correct independently (i.e., each commit should be linted with helm and pass tests).

- If your patch is not getting reviewed, or you need a specific person to review
  it, you can @-reply a reviewer asking for a review in the merge request or a
  comment.

- Add tests relevant to the fixed bug or new feature.

## Development tools

- [Helm](https://helm.sh/docs/intro/install/) linting and templating

  > ```sh
  > helm lint .
  > helm template .
  > ```

- [Go](https://golang.org/doc/install) creating and running tests
  > use `go test ./... -v -run TestName` to run a single test
- [gotestsum](https://github.com/gotestyourself/gotestsum) prints formatted test output,
  and a summary of the test run
  > `gotestsum --format testname`
- [helm-docs](https://github.com/norwoodj/helm-docs) generates documentation
  > `helm-docs --template-files=./docs/README.md.gotmpl --template-files=./docs/_templates.gotmpl --chart-search-root=..`

## Example development cycle

1. one of the next:
   - create or modify test(s) for your new feature in ***chart-test***
   - modify the ***helm-common*** templates
1. package helm-common chart (if you are in the chart-test folder `helm package ../`)
1. move the packaged chart to the chart-test/charts folder (if you are in the chart-test folder `mv -v helm-common-* charts/`)
1. run the tests (if you are in the chart-test folder `gotestsum --format testname`)
   - if all tests are green then maybe you are good to go
   - else back to Step One

> Step 2-4. are in the **chart-test/run-tests.sh** so you only need to continuously run it, and watch the output,
> so you can focus on step one. (For example: `watch -n 10 ./run-tests.sh`)

It's strongly advised to try it out with a local kubernetes installation before committing/pushing.
(You can install locally and set parameters for the chart-test chart)

## Conduct

We follow the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
