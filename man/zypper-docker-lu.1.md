% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% JUNE 2018
# NAME
zypper\-docker list-updates \- List all the available updates.

zypper\-docker list-updates-container \- List all the available updates for
the given container.

# SYNOPSIS
**zypper-docker list-updates** IMAGE

**zypper-docker list-updates-container** CONTAINER

# DESCRIPTION
The **list-updates** command lists all the updates that are available for the
given openSUSE/SUSE Linux Enterprise image. The provided image follows the
same naming conventions as in Docker. To fetch which images are based on
openSUSE or SUSE Linux Enterprise, use the **images** command.

The **list-updates-container** command takes the container ID and lists the updates for
the given container. Note that **list-updates-container** will not modify a running
container. Instead of that, **zypper-docker** will spawn a new container that will
then be analyzed. **List-updates-container** is also able to analyze stopped containers.
The **--base** flag can be used to analyze the base image of the container instead.

# COMMAND OPTIONS
**--base**
  Analyze the base image of the container for updates.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
June 2018, updated for v1.3.0 by Pascal Arlt <parlt@suse.com>
