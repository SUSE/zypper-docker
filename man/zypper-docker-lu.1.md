% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% SEPTEMBER 2015
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

The **list-updates-container** takes the container ID and lists the updates for
the image in which the given container is based on. Note that
**list-updates-container** will not modify a running container. Instead of
that, **zypper-docker** will spawn a new container based on the image in which
the running container is based on.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
