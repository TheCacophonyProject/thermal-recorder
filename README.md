# thermal-recorder

This software is used by The Cacophony Project to record thermal video
footage from a FLIR Lepton 3 camera when a warm moving object
(hopefully an animal) is detected. Recordings are stored using the
project's own CPTV format.

## Releases

Releases are built using TravisCI. To create a release:

* Tag the release with an annotated tag. For example:
  `git tag -a "v1.4" -m "1.4 release"`
* Push the tag to Github: `git push origin v1.4`
* TravisCI will see the pushed tag, run the tests, create a release
  package and create a
  [Github Release](https://github.com/TheCacophonyProject/thermal-recorder/releases).

For more about the mechanics of how releases work, see `travis.yml`
and `.goreleaser.yml`.
