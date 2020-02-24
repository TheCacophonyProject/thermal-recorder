# thermal-recorder

This software is used by The Cacophony Project to record thermal video
footage from FLIR Lepton 3 and Boson cameras when a warm moving object
(hopefully an animal) is detected. Recordings are stored using the
project's own CPTV format.

## Releases

Releases are built using TravisCI. To create a release visit the
[repository on Github](https://github.com/TheCacophonyProject/thermal-recorder/releases)
and then follow our [general instructions](https://docs.cacophony.org.nz/home/creating-releases)
for creating a release.

For more about the mechanics of how releases work, see `.travis.yml` and `.goreleaser.yml`.

## thermal-writer

The thermal-writer tool included in this codebase is designed for
specialised sitations where continous capture of thermal video frames
is required. It is intended to be run instead of thermal-recorder.

When recording Boson 640 frames at 60Hz, the following configuration
is recommended:

- Raspberry Pi 4
- External storage (e.g. USB SSD)
- Use the `performance` CPU scaling governor (write `performance` to `/sys/devices/system/cpu/cpufreq/policy0/scaling_governor`)
