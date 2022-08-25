# Contributing to hcdiag
We hope that hcdiag makes collecting diagnostic details about HashiCorp products as easy as possible, so that our
customers have as seamless an experience as possible when they need support. To that end, we are always looking for ways
to improve the tool! If you have ideas for improvements, if you’ve found bugs in the tool, or if you would like to add
to the codebase, we welcome your contributions!

The purpose of this document is to lay out guidelines for contributing to the project.

## Feature Requests
If you have an idea for a new capability that you think would improve hcdiag, please let us know! The proper channel for
these recommendations is via a GitHub Issue. Please visit [this page][gh-issues] and click the green “New Issue” button.
From there, select “Feature request,” and fill out the form, following the prompts provided. Feature requests will be
added to the product discussion backlog for prioritization. The GitHub issue that is created will be used for
communication about the request.

## Bug Reports
If you have found some behavior in hcdiag that isn’t quite what you expected, please let us know by filing a Bug Report!
Like with Feature Requests, we track Bug Reports in GitHub. Please visit [this page][gh-issues] and click the green
“New Issue” button. From there, select “Bug report,” and fill out the form, following the prompts provided. As much
detail as you can provide will help us more quickly identify the issue, so please describe the issue as thoroughly as
you’re able to. Like with Feature Requests, communication will be handled by way of the GitHub issue that you create,
but we will prioritize any known bugs as quickly as possible.

## Security Vulnerabilities
If you’ve identified a suspected security vulnerability, we would appreciate responsible disclosure. Please see the
[policy document here][security-policy] for more information about where to report security concerns about HashiCorp
products.

## Making Code Contributions
We welcome code contributions from collaborators both inside and outside HashiCorp! The process is generally the same
for both sets of contributors, and it is detailed below, with any differences noted.

### GitHub Issue
The first step is to ensure there is a GitHub Issue that captures the essence of what you want to provide. If you have
an idea for a new feature or have identified an unknown bug, please use the guidelines above in order to file an
appropriate issue.

Please add a comment that you would like to work on a PR to address the issue yourself, so we can respond quickly to
confirm that the capability is something that fits into the long-term vision of hcdiag.

In general, if it makes it easier to use the tool or to get relevant information about HashiCorp products, it makes
sense! However, there are also ancillary data gathering use cases, like metrics collection, which we do not plan to add
to hcdiag. We don’t want to make for a frustrating experience where you create a pull request that we ultimately
cannot accept.

### Creating a Working Branch
Once you’ve received confirmation that an issue is ready for you to work on, you’re able to get started whenever you
want. Please create a branch on your local machine off of the latest version of hcdiag’s `main` branch. Here is where
there is a slight difference between internal and external contributors.

If you do not work at HashiCorp, you will need to fork the repository and subsequently open a PR from your repo to ours.

If you do work at HashiCorp, you should have permission to create development branches. So, you should be able to simply
clone the repo and create a branch. We prefer branch names that include your GitHub username or some way to identify you
in case we have questions. This is not enforced, but we do ask for your consideration here.

### Make Code Changes
The fun part!

In addition to making the relevant code changes, please also include unit tests for any areas of the code that you
touch. For new features, unit tests help us verify that certain scenarios have been considered, and they generally
provide a great example of how the person who added the capability actually expects it to work. And, for bugs, unit
tests help us better ensure that we don’t introduce regressions in the future. There may be cases where tests are
infeasible, but if you feel that this is the case in your code, please be aware that we will ask about it in PR reviews.

### Pull Requests
When you think your code is ready for merging, it’s time to open a Pull Request! (Note, if you don’t think your code is
QUITE ready for merging, but you’d like to solicit feedback, feel free to open a Draft Pull Request instead!)

Before opening the PR, please make sure that your branch is able to be merged into main. We suggest that you pull the
latest version of main and then rebase your changes onto it.

In the PR description, please provide a synopsis of what a user “gets” with this code change - what is the functionality
that has been added or fixed? Also, please link it to the GitHub issue associated with the feature or bug.

The target for PRs should be the `main` branch. The hcdiag maintainers will take care of back-porting changes to feature
branches if needed.

#### PR Changelog Entries
One small “surprise” that you may run into with pull requests is that we rely on files checked into the repo to generate
our changelogs. There needs to be a file in the `changelog` directory with the name `<your PR #>.txt`. So, in the case
of a pull request that was given number 209, the file would be `209.txt`. Without a file that is named properly, the CI
pipeline will fail the PR. Unfortunately, there is a bit of “chicken-and-egg” here because you won’t know the PR number
for sure until you create it. So, please forgive the slight hiccup in the process where you’ll need to create the PR,
then go back to your local branch to put the proper file in place, and then push your changes back up. (These changes
will automatically kick off another changelog check, which should pass this time.)

The format for these files is described in the file `changelog/README.md`. Please refer to that document and review some
existing files for some ideas on how to fill in the changelog file.

### Pull Request Reviews
The hcdiag maintainers will review PRs and respond with questions or concerns about the submitted code. Expect that
there could be some back-and-forth as we try and make sure we understand and are comfortable merging your changes.

### Merging
Merging PRs is done by hcdiag maintainers. Once we have approved the PR and we know that you are comfortable with it, we
will merge your changes!

[gh-issues]:        https://github.com/hashicorp/hcdiag/issues
[security-policy]:  https://github.com/hashicorp/hcdiag/security/policy