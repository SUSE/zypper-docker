% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% JUNE 2018
# NAME
zypper\-docker list-patches \- List all the available patches.

zypper\-docker list-patches-container \- List all the available patches for
the given container.

# SYNOPSIS
**zypper-docker list-patches** [command options] IMAGE

**zypper-docker list-patches-container** [command options] CONTAINER

# DESCRIPTION
The **list-patches** command lists all the patches that are available for the
given openSUSE/SUSE Linux Enterprise image. The provided image follows the
same naming conventions as in Docker. To fetch which images are based on
openSUSE or SUSE Linux Enterprise, use the **images** command.

The **list-patches-container** takes the container ID and lists the patches for
the given container. Note that **list-patches-container** will not modify a running
container. Instead of that, **zypper-docker** will spawn a new container that will
then be analyzed. **List-patches-container** is also able to analyze stopped containers.
The **--base** flag can be used to analyze the base image of the container instead.

# COMMAND OPTIONS
**--base**
  Analyze the base image of the container for patches.

**--bugzilla[=#bug-id]**
  List available needed patches for all Bugzilla issues, or issues whose number matches the given string (--bugzilla=#).

**--cve[=#cve-id]**
  List available needed patches for all CVE issues, or issues whose number matches the given string (--cve=#).

**--date**
  List patches issued up to, but not including, the specified date (YYYY-MM-DD).

**--issues**
  Look for issues whose number, summary, or description matches the specified string (--issue=string).

**-g**, **--category**
  List only patches with this category.

**--severity**
  List only patches with this severity. Note that this requires zypper >= 1.12.6 inside of your docker image.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
June 2018, updated for v1.3.0 by Pascal Arlt <parlt@suse.com>
