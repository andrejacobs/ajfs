# ajfs - Specification

Author: André Jacobs
Date: 3/10/2024

## Overview

Why mk2 (mark 2)?

-   About a year ago I wrote `ajfs` which I used internally quite a bit.
-   I now consider that project as the mk1 that helped me prototype and play around with ideas.
-   I learned a lot about what I would like to improve in the next version, which now brings me to mk2.

What is this tool for?

-   Used to take snapshots of file hierarchies, meta data and checksums.
-   Can be used to track differences. Between different hierarchies, different periods in time.
-   Can be used to browse, search etc. offline from the host on which the files exists.
-   Can be used to see if files have been changed (damaged maybe? need to be backed up maybe? etc.)

Who is it for?

-   The audience is mainly going to be me to start with because it solves a problem I actually have. Maybe others will find this useful too. But there will have to be big disclaimers about using this at own risk. Need to fully establish liability is not with me.

Why build this yourself?

-   I have been needing and thinking about a tool like this for a number of years now but never quite found something that I would use consistently.
    Most of the time it would be cobbled toghether shell scripts and 3rd party tools.
    There was also the mk1 version but I am not quite happy about the UX.

## Command summary

Overview of the subcommands that will be available.

### Core features

-   `scan`

    -   Build up a tree of directories and files
    -   Calculate hashes etc.
    -   Store the information in a single file "database".
    -   Will not update an existing database. Use `update` command for this.
