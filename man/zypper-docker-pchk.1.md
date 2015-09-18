% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% SEPTEMBER 2015
# NAME
zypper\-docker pchk \- Check for patches.

zypper\-docker patch-check-container \- Check for patches available for the
given container.

# SYNOPSIS
**zypper-docker patch-check** IMAGE

**zypper-docker patch-check-container** CONTAINER

# DESCRIPTION
The **patch-check** command checks for patches that are available for the
given openSUSE/SUSE Linux Enterprise image. The provided image follows the
same naming conventions as in Docker. To fetch which images are based on
openSUSE or SUSE Linux Enterprise, use the **images** command.

The **patch-check-container** takes the container ID and lists the patches for
the image in which the given container is based on.

# EXIT CODES
The **patch-check** command respects the same exit codes as provided by
**zypper**. In particular, for this command there are the following available
exit codes:

**0 \- ZYPPER\_EXIT\_OK**
  Successful run with no special info.

**100 \- ZYPPER\_EXIT\_INF\_UPDATE\_NEEDED**
  There are patches available for installation.

**101 \- ZYPPER\_EXIT\_INF\_SEC\_UPDATE\_NEEDED**
  There are security patches available for installation.

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
