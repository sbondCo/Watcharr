We don't really use the `internal` folder because none of our code is anything we'd ever expect anyone to import and rely on in their own codebase.

However, for some packages, it might make sense to put it here just as a signal to us (while developing) that this package is not code that goes into a prod build, etc (eg: `testutil`). This feels nicer, avoiding a scenario where our root folder has a bunch of real packages and ones that should never see prod mixed together (which is probably confusing).
