% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% JUNE 2018
# NAME
zypper\-docker \- Patching Docker images safely

# SYNOPSIS
**zypper-docker** [global options] command [command options] [arguments...]

# DESCRIPTION
**zypper-docker** provides a quick way to patch and update Docker Images based
on either SUSE Linux Enterprise or openSUSE.

**zypper-docker** mimics the zypper command line syntax to ease its usage.
This application relies on zypper to perform the actual operations against
Docker images.

**zypper-docker** has 11 different commands, all of them listed below in the
**COMMANDS** section. Moreover, each command has its own man page which
explains its usage and options. To read the man page of a specific command,
just run **man zypper-docker <command>**.

# GLOBAL OPTIONS
**-f**, **--force**
  zypper\-docker caches data that is expensive to compute into a local file.  This option forces zypper\-docker to ignore this cache file.

**--gpg-auto-import-keys**
  If a new repository signing key is found, do not ask what to do; trust and import it automatically

**--help**, **-h**
  Show the help message.

**--no-gpg-checks**
  Ignore GPG check failures and continue

**--add-host**
  You can specify has many additional hosts:ip mappings for the created containers.

**--version**, **-v**
  Print the version.

# COMMANDS
**images**
  List all the images based on either OpenSUSE or SUSE Linux Enterprise.
  See **zypper-docker-images(1)** for full documentation on the **images** command.

**list-updates**, **lu**
  List all the available updates.
  See **zypper-docker-list-updates(1)** or **zypper-docker-lu(1)** for full documentation on the **list-updates** command.

**list-updates-container**, **luc**
  List all the available updates for the given container.
  See **zypper-docker-list-updates(1)** or **zypper-docker-lu(1)** for full documentation on the **list-updates-container** command.

**update**, **up**
  Install the available updates.
  See **zypper-docker-update(1)** or **zypper-docker-up(1)** for full documentation on the **update** command.

**list-patches**, **lp**
  List all the available patches.
  See **zypper-docker-list-patches(1)** or **zypper-docker-lp(1)** for full documentation on the **list-patches** command.

**list-patches-container**, **lpc**
  List all the available patches for the given container.
  See **zypper-docker-list-patches(1)** or **zypper-docker-lp(1)** for full documentation on the **list-patches-container** command.

**patch**
  Install the available patches.
  See **zypper-docker-patch(1)** for full documentation on the **patch** command.

**patch-check**, **pchk**
  Check for patches.
  See **zypper-docker-patch-check(1)** or **zypper-docker-pchk(1)** for full documentation on the **patch-check** command.

**patch-check-container**, **pchkc**
  Check for patches available for the given container.
  See **zypper-docker-patch-check(1)** or **zypper-docker-pchk(1)** for full documentation on the **patch-check-container** command.

**ps**
  List all the containers that are outdated.
  See **zypper-docker-ps(1)** for full documentation on the **ps** command.

**help**, **h**
  Shows a list of commands or help for one command.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
June 2018, updated for v2.0.0 by Pascal Arlt <partl@suse.com>
