# elect
Simple, by-the-book implementation of Raft consensus

** WORK-IN-PROGRESS **

The ultimate goal is to produce an implementation that works, has been tested, and is understandable.

## Instructions

#### Download the code
```
go get -d github.com/kalexmills/elect
```

#### Install the CLI
```
go install github.com/kalexmills/elect
```

#### Run a local demo
```
elect launch cluster 1234 2345 3456
```

The above commandd launches three subprocesses, each listening on ports 1234 2345 3456, respectively. The processes immediately begin running the Raft protocol, logging to STDOUT. Killing the main process will cleanup the subprocesses.
