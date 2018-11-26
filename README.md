# thermal-recorder

This software is used by The Cacophony Project to record thermal video
footage from a FLIR Lepton 3 camera when a warm moving object
(hopefully an animal) is detected. Recordings are stored using the
project's own CPTV format.

## Releases

The software uses the [GoReleaser](https://goreleaser.com) tool to
automate releases. To produce a release:

* Tag the release with an annotated tag. For example:
  `git tag -a "v1.4" -m "1.4 release"`
* Push the tag to Github: `git push origin v1.4`
* Travis should run automatically and create the release.

The configuration for GoReleaser can be found in `.goreleaser.yml`.
