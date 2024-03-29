# Bigquery Dataset Tools

These tools help us look at our tables.

Requirements:

* bigquery access and credential files (the ones in the make file are user specific so you
  need to modify those).

## How to build and run

This section shows how to build the log files which contain the names of the datasets and tables
sorted by "last modified".  I also included information about whether the table is partitioned
and the list of columns in case we're interested in what other things we store.

```bash

# Make the executable
make build

# Make the output file for openshift-ci-data-analysis.
make gen1

# Make remove the table column names from the log file.
make gen1a

# Make the output file for openshift-gce-devel.
make gen2

# Make remove the table column names from the log file.
make gen2a

# Make all of the above.
#
make all

# Cleanup everything.
make clean
```