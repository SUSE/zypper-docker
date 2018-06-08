% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% JUNE 2018
# NAME
zypper\-docker ps \- List all the running containers that are outdated.

# SYNOPSIS
**zypper-docker ps**

# DESCRIPTION
The **ps** command goes through all the running containers and lists which one
of them are based on an outdated openSUSE/SUSE Linux Enterprise image. In order
to detect which containers are based on an outdated image, **zypper-docker**
takes into account the history of patched images. Therefore, do not expect
the **ps** command to provide feedback about *all* the possible SUSE
containers.

In order to properly detect whether a specific running container is outdated or
not, use either the **list-patches-container** or the **patch-check-container**
commands.

This command does not accept any extra arguments or command options.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
June 2018, updated for v2.0.0 by Pascal Arlt <partl@suse.com>
