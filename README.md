## zypper-docker [![Build Status](https://travis-ci.org/SUSE/zypper-docker.svg?branch=master)](https://travis-ci.org/SUSE/zypper-docker) [![GoDoc](https://godoc.org/github.com/SUSE/zypper-docker?status.png)](https://godoc.org/github.com/SUSE/zypper-docker)

The `zypper-docker` command line tool provides a quick way to patch and update
Docker Images based on either SUSE Linux Enterprise or openSUSE.

`zypper-docker` mimics `zypper` command line syntax to ease its usage. This
application relies on `zypper` to perform the actual operations against Docker
images.

[![asciicast](https://asciinema.org/a/26248.png)](https://asciinema.org/a/26248)

## Targetting Docker daemons running on remote machines

`zypper-docker` can interact with docker daemons running on remote machines.
To do that it uses the same environment variables of the
[docker client](https://docs.docker.com/reference/commandline/cli/#environment-variables).

[docker-machine](https://docs.docker.com/machine/) can be used to configure the
remote Docker host and setup the local environment variables.

[![asciicast](https://asciinema.org/a/26244.png)](https://asciinema.org/a/26244)

## Commands

### Listing images

The **images** command is similar to the one from Docker, but this one only
lists images that are based on either openSUSE or SUSE Linux Enterprise. Here's
an example:

```
mssola:~ $ docker images
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
opensuse            latest              c7ff47bc7ebb        13 days ago          254.5 MB
busybox             latest              8c2e06607696        3 months ago         2.43 MB
mssola:~ $ zypper-docker images
REPOSITORY          TAG                 IMAGE ID            CREATED              VIRTUAL SIZE
opensuse            latest              c7ff47bc7ebb        13 days ago          254.5 MB
```

### Updates

First of all, you can check whether an image has pending updates or not by
using the **list-updates** command. The usage is as follows:

```
$ zypper docker list-updates (lu) <image>
```

Similarly, there is the **list-updates-container** that does the same but
targeting an already running container. Note that this command does *not* touch
the running container, but it just detects the image in which the running
container is based on, and then it just performs **list-updates** for the
according image. There's a short video about **list-updates** in action here:

[![asciicast](https://asciinema.org/a/25310.png)](https://asciinema.org/a/25310)

But more important than listing updates is to actually install them. You can do
this with the **update** command. It has the following usage:

```
$ zypper docker update (up) [options] <image> <new-image>
```

If there are updates, this command will create a new Docker image based on
the given image, but with the needed updates already installed. Therefore, note
that `zypper-docker` will *never* change anything from the old image. More than
that, this command will refuse to overwrite an already existing Docker image.

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
* `--author`: commit author to associate with the new layer. By default it
  uses the canonical name of the current user.
* `--message`: commit message to be associated with the new layer. If no
  message was provided, zypper-docker will write: "[zypper-docker] update".

You can find a small video about the **update** Command here:

[![asciicast](https://asciinema.org/a/25312.png)](https://asciinema.org/a/25312)

### Patches

The operations that can be done for patches is very similar to the ones that
can be done for updates. Therefore, the **list-patches** and the
**list-patches-container** commands are almost identical to the ones for
updates. In particular, these are their usage:

```
$ zypper docker list-patches (lp) [options] image

$ zypper docker list-patches-container (lpc) [options] image
```

As you will notice, both of these commands accept some options. For these
commands the user can filter the results according to the following attributes:

The available options are:
* `--bugzilla[=#]`: List available needed patches for all Bugzilla issues,
  or issues whose number matches the given string.
* `--cve[=#]`: List available needed patches for all CVE issues, or issues
  whose number matches the given string.
* `--date YYYY-MM-DD`: List patches issued up to, but not including, the
  specified date.
* `--issues[=string]`: Look for issues whose number, summary, or description
  matches the specified string. Issues found by number are displayed
  separately from those found by descriptions. In the latter case, use zypper
  patch-info patchname to get information about issues the patch fixes.
* `-g, --category category`: List available patches in the specified category.

You can find a small video on listing patches here:

[![asciicast](https://asciinema.org/a/25311.png)](https://asciinema.org/a/25311)

Interestingly enough, `zypper` can also just check whether there are patches
available at all. This is really convenient for using `zypper` inside of
scripts. `zypper-docker` also implements this in the form of the
**patch-check** command. This is its usage:

```
$ zypper docker patch-check (pchk) image
```

This command will exit with a status code of **100** if there are patches
available, and **101** if there are not.

Besides listing and checking for patches, you can also of course install them.
You do that with the **patch** command. It has the following usage:

```
$ zypper docker patch [options] image new-image
```

Similarly to the **update** command, this command will not change the original
change, but it creates a new patched image. This command also takes into
account that the new image does not overwrite an already existing one. The
arguments that can be passed to this command are as follows:

* `--bugzilla #`: Install patch fixing a Bugzilla issue specified by
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
* `--author`: commit author to associate with the new layer. By default it
  uses the canonical name of the current user.
* `--message`: commit message to be associated with the new layer. If no
  message was provided, zypper-docker will write: "[zypper-docker] patch".

You can find a small video showing off the **patch** command here:

[![asciicast](https://asciinema.org/a/25315.png)](https://asciinema.org/a/25315)

### List all the missing updates

Lastly, `zypper-docker` also has the **ps** command. This command traverses
through all the running containers and investigates which of them are based on
images that have been recently upgraded. Therefore, this command does *not*
provide feedback about *all* the possible SUSE containers, only the ones that
have been updated/patched with the **update** and **patch** commands.

This command doesn't have any options, so the usage is quite straight-forward:

```
$ zypper docker ps
```

## Local cache

Note that some of these commands might be expensive. That's why some of the
needed data is cached into a single file. This file is named
`docker-zypper.json`. This cache file normally resides inside of the
`$HOME/.cache` directory. However, if there is some problem with this
directory, it might get saved inside of the `/tmp` directory.

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

### Run the tests

To run the test suite and the code analysis tools against all Go versions,
type:

```
$ make test
```

### Integration tests

The integration tests invoke the `zypper-docker` binary and test different
scenarios. They are written using RSpec and are located under `/spec`. The
integration tests can be started by doing:

```
make test_integration
```

This will build a Docker image containing all the software (RSpec, plus other
Ruby gems) required to run the tests. The image will be started, the socket
used by the Docker daemon on the host will be mounted inside of the new
container. That makes possible to invoke the docker client from within the
container itself.

## License

Licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/SUSE/zypper-docker/blob/master/LICENSE) for
the full license text.
