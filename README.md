# thermal-recorder

This software is used by The Cacophony Project to record thermal video
footage from a FLIR Lepton 3 camera when a warm moving object
(hopefully an animal) is detected. Recordings are stored using the
project's own CPTV format.

## Releases

The software uses the [GoReleaser](https://goreleaser.com) tool to
automate releases. To produce a release:

* Ensure that the `GITHUB_TOKEN` environment variable is set with a
  Github personal access token which allows access to the Cacophony
  Project repositories.
* Tag the release with an annotated tag. For example:
  `git tag -a "v1.4" -m "1.4 release"`
* Push the tag to Github: `git push --tags origin`
* Run `goreleaser`

The configuration for GoReleaser can be found in `.goreleaser.yml`.
