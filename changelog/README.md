# changelog

Release notes are text files with three lines:

1. An opening code block with the `release-note:<MODE>` type annotation.

   For example:

       ```release-note:bug

   Valid modes are:

    - `bug` - Any sort of non-security defect fix.
    - `security` - Any fixes for security issues in previous versions.
    - `change` - A change in the product that may require action or
      review by the operator. Examples would be any kind of API change
      (as opposed to backwards compatible addition), a notable behavior
      change, or anything that might require attention before updating. Go
      version changes are also listed here since they can potentially have
      large, sometimes unknown impacts. (Go updates are a special case, and
      dep updates in general aren't a `change`). Discuss any potential
      `change` items in the pull request to see what other communication
      might be warranted.
    - `deprecation` - Announcement of a planned future removal of a
      feature.
    - `feature` - Large topical additions for a major release. These are
      rarely in minor releases. Formatting for `feature` entries differs
      from normal changelog formatting - see the [new features
      instructions](#new-and-major-features).
    - `improvement` - Most updates to the product that aren’t `bug`s, but
      aren't big enough to be a `feature`, will be an `improvement`.

3. A component (for example, `seeker` or `agent`), a colon and a space,
and then a one-line description of the change.

4. An ending code block.

This should be in a file named after the pull request number (e.g., `12345.txt`).

See [hashicorp/go-changelog](https://github.com/hashicorp/go-changelog) for full documentation on the supported entries.

## New and Major Features

For features that we are introducing in a new major release, we prefer a single
changelog entry representing that feature. This way, it is clear to readers
what feature is being introduced. You do not need to reference a specific PR,
and the formatting is slightly different - your changelog file should look
like:

    changelog/<pr num OR feature name>.txt:
    ```release-note:feature
    **Feature Name**: Description of feature - for example "Custom password policies are now supported for all database engines."
    ```
