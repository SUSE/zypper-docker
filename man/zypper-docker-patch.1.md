% ZYPPER-DOCKER(1) zypper-docker User manuals
% SUSE LLC.
% SEPTEMBER 2015
# NAME
zypper\-docker patch \- Install the available patches for the given image.

# SYNOPSIS
**zypper-docker patch** [command options] IMAGE NEW-IMAGE

# DESCRIPTION
The **patch** command patches the given openSUSE/SUSE Linux Enterprise image
with all the available updates. The updated image will have a new name, as
provided by the NEW-IMAGE argument.

To list all the patches available for a given image, use the **list-patches**
and the **list-patches-container** commands. To show all the images based on
openSUSE/SUSE Linux Enterprise, use the **images** command.

# COMMAND OPTIONS
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
