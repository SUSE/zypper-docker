## zypper-docker [![Build Status](https://travis-ci.org/SUSE/zypper-docker.svg?branch=master)](https://travis-ci.org/SUSE/zypper-docker) [![GoDoc](https://godoc.org/github.com/SUSE/zypper-docker?status.png)](https://godoc.org/github.com/SUSE/zypper-docker)

The `zypper-docker` command line tool provides a quick way to patch and update
Docker Images based on either SUSE Linux Enterprise or openSUSE.

`zypper-docker` mimics `zypper` command line syntax to ease its usage. This
application relies on `zypper` to perform the actual operations against Docker
images.

[![asciicast](https://asciinema.org/a/25309.png)](https://asciinema.org/a/25309)

**NOTE**: this application is still WIP. Here's a list of the features that
have been implemented:

- [ ] Global options should be respected. (the implemented commands do it)
- [x] List all the available images.
- [x] List all the available updates.
- [x] List all the available patches.
- [x] Checking patches.
- [x] Installing patches.
- [x] Install updates.



## Generic operations

This tool supports some of the global options as defined by zypper. They are
all set to false by default:

* `-n`, `--non-interactive`
* `--no-gpg-checks`
* `--gpg-auto-import-keys`
* `-f`, `--force`: ignore cached values.

Note that some of these commands might be expensive. That's why some of the
needed data is cached into a single file. This file is named
`docker-zypper.json` and it can be located in either of these locations:

1. $XDG\_CACHE\_HOME
2. $XDG\_DATA\_DIRS
3. $HOME/.cache
4. /tmp

The application will first try to allocate the cache on `$XDG_CACHE_HOME`. If
it fails, it will try it on the next location, and so on.

### List all the available images:

List all the Docker images that are based on openSUSE or SLE (that is, any
system with the `zypper` command installed). Here's an example:

```
mssola:~ $ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
opensuse            latest              c7ff47bc7ebb        13 days ago          254.5 MB
busybox             latest              8c2e06607696        3 months ago         2.43 MB
mssola:~ $ zypper-docker images
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
opensuse            latest              c7ff47bc7ebb        13 days ago          254.5 MB
```

## Operations available against Docker images

**TODO**

  * Specify the name of the final image `-t` like `docker build` maybe?
  * Should we handle non-interactive stuff?


### List all the updates available

We can list the updates available with the following command:

```
$ zypper docker list-updates (lu) [options] image
```

Note that even if zypper supports some options, we don't because they do not
really apply to this tool.

[![asciicast](https://asciinema.org/a/25310.png)](https://asciinema.org/a/25310)

### Install updates

Install all available updates.

```
$ zypper docker update (up) [options] image target
```

The command will create a new Docker image as specified by `target` (e.g.: `flavio/redis:1.1.0`).

The command will refuse to overwrite an already existing Docker image.

The available options are:

* `--skip-interactive`: This will skip interactive patches, that is, those
  that need reboot, contain a message, or update a package whose license needs
  to be confirmed.
* `--with-interactive`: Avoid skipping of interactive patches when in
  non-interactive mode.
* `-l, --auto-agree-with-licenses`: Automatically say yes to third party
  license confirmation prompt. By using this option, you choose to agree with
  licenses of all third-party software this command will install.
  This option is particularly useful for administrators installing the same
  set of packages on multiple machines (by an automated process) and have the
  licenses confirmed before.
* `--no-recommends`: By default, zypper installs also packages recommended by
  the requested ones. This option causes the recommended packages to be
  ignored and only the required ones to be installed.
* `--replacefiles`: Install the packages even if they replace files from
  other, already installed, packages. Default is to treat file conflicts as an
  error.

[![asciicast](https://asciinema.org/a/25312.png)](https://asciinema.org/a/25312)

### List patches available

We can list the patches available with the following command:

```
zypper docker list-patches (lp) [options] image
```

The available options are:
* `-b, --bugzilla[=#]`: List available needed patches for all Bugzilla issues,
  or issues whose number matches the given string.
* `--cve[=#]`: List available needed patches for all CVE issues, or issues
  whose number matches the given string.
* `--date YYYY-MM-DD`: List patches issued up to, but not including, the
  specified date.
* `-g, --category category`: List available patches in the specified category.
* `--issues[=string]`: Look for issues whose number, summary, or description
  matches the specified string. Issues found by number are displayed
  separately from those found by descriptions. In the latter case, use zypper
  patch-info patchname to get information about issues the patch fixes.

[![asciicast](https://asciinema.org/a/25311.png)](https://asciinema.org/a/25311)

### Check for patches

Check for patches. Displays a count of applicable patches and how many of them
have the security category.

### Install patches

Install all available needed patches.

```
zypper docker patch [options] image
```

**NOTE WELL**
If there are patches that affect the package management itself, those will be
installed first and you will be asked to run the patch command again.


Options to port:
  * `-b, --bugzilla #`: Install patch fixing a Bugzilla issue specified by
    number. Use list-patches --bugzilla command to get a list of available
    needed patches for specific issues.
  * `--cve #`: Install patch fixing a MITRE’s CVE issue specified by number.
    Use list-patches --cve command to get a list of available needed patches for
    specific issues.
  * `--date YYYY-MM-DD`: Install patches issued up to, but not including, the
    specified date.
  * `-g, --category category`: Install all patches in the specified category.
    Use list-patches --category command to get a list of available patches for
    a specific category.
  * `--skip-interactive`: Skip interactive patches.
  * `--with-interactive`: Avoid skipping of interactive patches when in
    non-interactive mode.
  * `-l, --auto-agree-with-licenses`: See the update command for description of
    this option.
  * `--no-recommends`: By default, zypper installs also packages recommended by
    the requested ones. This option causes the recommended packages to be
    ignored and only the required ones to be installed.
  * `--replacefiles`: Install the packages even if they replace files from
    other, already installed, packages. Default is to treat file conflicts as an
    error.
  * `--download-as-needed`: disables the fileconflict check because access tos
     all packages filelists is needed in advance in order to perform the check.

Options to drop:
  * `-r, --repo alias|name|#|URI`: Work only with the repository specified by
    the alias, name, number, or URI. This option can be used multiple times.

Options we might drop:
  * `--details`: Show the detailed installation summary.

**TODO investigate:** This command also accepts the download-and-install mode options described in the install command description.
**TODO:** handle interactive mode

[![asciicast](https://asciinema.org/a/25315.png)](https://asciinema.org/a/25315)

## Operations available against Docker containers

### List all the missing updates:

List all the containers that are based on an image recently upgraded by
zypper-docker.

```
zypper docker ps
```

## Development environment

It is possible to run all the test suite and the code analysis tool using
docker.

### Build the docker images

The tests and code analysis tool are going to be executed ran using both
the latest stable version of Go and the current experimental version.

To build these docker images type:

```
$ make build
```

### Run the tests against Go stable

To run the test suite and the code analysis tools against Go stable type:

```
$ make test_stable
```

### Run the tests against Go tip

To run the test suite and the code analysis tools against Go tip type:

```
$ make test_tip
```

### Run the tests against Go stable and Go tip

To run the test suite and the code analysis tools against Go stable and Go tip
type:

```
$ make test
```

## Testing

This project is covered both by unit and integration tests.

### Unit tests

Unit tests are written in Go and can be invoked by doing:

```
make test
```

This will build `zypper-docker` using different Go versions and trigger the unit tests.

### Integration tests

The integration tests invoke the `zypper-docker` binary and test different scenarios.

They are written using RSpec and are located under `/spec`.

The integration tests can be started by doing:

```
make test_integration
```

This will build a Docker image containing all the software (RSpec, plus other
Ruby gems) required to run the tests.
The image will be started, the socket used by the Docker daemon on the host will
be mounted inside of the new container. That makes possible to invoke the docker
client from within the container itself.

## License

Licensed under the Apache License, Version 2.0. See
[LICENSE](https://gitlab.suse.de/docker/zypper-docker/blob/master/LICENSE) for
the full license text.
