## Preamble

This repository is a temporary placeholder. Once everything is settled we will
move it to github.com under the SUSE umbrella.

First of all we need to decide how the whole workflow is going to look like,
then we will implement the actual command.


The `zypper-docker` command line tool provides a quick way to patch and update
Docker Images based either on SUSE Linux Enterprise or openSUSE.

`zypper-docker` mimics `zypper` command line syntax to ease its usage.

`zypper-docker` relies on `zypper` to perform the actual operations against
Docker images.


## Generic operations

**TODO:** handle generic options of zypper like:
  * `-n`: non interactive
  * `-i`: ignore unknown packages
  * `--no-gpg-checks`
  * `--gpg-auto-import-keys`

### List all the available images:

List all the Docker images that are based on openSUSE or SLE (or simply have
`/usr/bin/zypper` inside of them).

```
$ zypper docker images
```

## Operations available against Docker images

**TODO**

  * Specify the name of the final image `-t` like `docker build` maybe?
  * Should we handle non-interactive stuff?


### List all the updates available

```
$ zypper docker list-updates (lu) [options] image
```

zypper options to drop:
  * `-t, --type type`: Type of package (default: package). See section Package
     Types for list of available package types. If patch is specified, zypper
     acts as if the list-patches command was executed.
  *  `-r, --repo alias|name|#|URI`: Work only with the repository specified by
     the alias, name, number, or URI. This option can be used multiple
     times.

zypper options we might consider to keep around:
  * `-a, --all`: List all packages for which newer versions are available,
    regardless whether they are installable or not.
  * `--best-effort`: See the update command for description.

### List patches available

List all available needed patches.

```
zypper docker list-patches (lp) [options] image
```

zypper options to keep:
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

zypper options to drop:
  * `-a, *--all`: By default, only patches that are relevant and needed on your
    system are listed. This option causes all available released patches to be
    listed. This option can be combined with all the rest of the list-updates
    command options.
  * `-r, --repo alias|name|#|URI`: Work only with the repository specified by
    the alias, name, number, or URI. This option can be used multiple times.

### Check for patches

Check for patches. Displays a count of applicable patches and how many of them
have the security category.

```
zypper docker patch-check (pchk) image
```

See also the EXIT CODES section for details on exit status of 0, 100, and 101
returned by this command.

**TODO: Add EXIT CODES section**

zypper options to drop:
  * `-r, --repo alias|name|#|URI`: Check for patches only in the repository
    specified by the alias, name, number, or URI. This option can be used
    multiple times.

### Install patches

Install all available needed patches.

```
zypper docker patchh [options] image
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

## Operations available against Docker containers

### List all the missing updates:

List all the containers that are based on an image recently upgraded by
zypper-docker.

```
zypper docker ps
```

