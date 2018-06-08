% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% JUNE 2018
# NAME
zypper\-docker update \- Install the available updates for the given image.

# SYNOPSIS
**zypper-docker update** [command options] IMAGE NEW-IMAGE

# DESCRIPTION
The **update** command updates the given openSUSE/SUSE Linux Enterprise image
with all the available updates. If there is a zypper-related update, which returns
the exit code 103 (ZYPPER_EXIT_INF_RESTART_NEEDED), zypper-docker goes through the 
update process once again, to install the remaining updates.
The updated image will have a new name, as provided by the NEW-IMAGE argument.
Note that both the IMAGE and the NEW\-IMAGE arguments have to follow Docker's naming format.

To list all the updates available for a given image, use the **list-updates**
and the **list-updates-container** commands. To show all the images based on
openSUSE/SUSE Linux Enterprise, use the **images** command.

# COMMAND OPTIONS
**-l**, **--auto-agree-with-licenses**
  Automatically say yes to third party license confirmation prompts. By using this option, you choose to agree with licenses of all third-party software this command will install.

**--no-recommends**
  By default, zypper installs also packages recommended by the requested ones. This option causes the recommended packages to be ignored and only the required ones to be installed.

**--replacefiles**
  Install the packages even if they replace files from other, already installed, packages. Default is to treat file conflicts as an error.

**--author**
  Commit author to associate with the new layer (e.g., \"John Doe <john.doe@example.com>\"). It defaults to the user's system login currently being used.

**--message**
  Commit message to associated with the new layer. If no message was provided, **zypper-docker** will write: "[zypper-docker] update".

# HISTORY
September 2015, created by Miquel Sabaté Solà <msabate@suse.com>
June 2018, updated for v2.0.0 by Pascal Arlt <partl@suse.com>
