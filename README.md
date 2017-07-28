# go-testdb
parses Golang test output and places relevant data into a MySql database

Currently it is specifically catered to the output of `make test-vlong` of Sia, but wouldn't take much effort to use with other Golang projects.
#### TODO
+ Add ability to query databases.

#### Table Setup

The `tests` table stores output for each test with the following fields (and corresponding types):
+ `commitHash`, `VARCHAR(40)`: commit hash of the head of the master branch of Sia at the time the test was run.
+ `dateTime`, `DATETIME`: date and time at which test was started in the format '2006-01-02-15:04:05'.
+ `name`, `VARCHAR(150)`: name of the test.
+ `result`, `ENUM('PASSED','SKIPPED','FAILED','UNDETERMINED')`: result of the test. A test is considered `UNDETERMINED` if it is started, but has no completion message. This can occur in the case where some other test causes a panic before it completes.
+ `output`,`TEXT`: the output of the test (e.g. the reason it was skipped or the reason it failed).
+ `duration`, `INT`: the duration of the test in seconds.

The `packages` table stores outputs that summarize the tests for an entire package with the following fields:
+ `commitHash`, `VARCHAR(40)`: commit hash of the head of the master branch of Sia at the time the packages tests was run.
+ `dateTime`, `DATETIME`: date and time at which test was started in the format '2006-01-02-15:04:05'.
+ `name`, `VARCHAR(150)`: name of the test.
+ `result`, `ENUM('PASSED','SKIPPED','FAILED','UNDETERMINED')`: result of the test. A test is considered `UNDETERMINED` if it is started, but has no completion message. This can occur in the case where some other test causes a panic before it completes.
+ `duration`, `INT`: the duration of the test in seconds.
