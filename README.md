# elect
Simple, [by-the-book](https://raft.github.io/raft.pdf) implementation of Raft consensus

** WORK-IN-PROGRESS **
Right now, leader election *seems* to be working as expected under normal circumstances.

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

The above command launches three subprocesses, each listening on ports 1234 2345 3456, respectively. The processes immediately begin running the Raft protocol, logging to STDOUT. Killing the main process will cleanup the subprocesses.

### Run a single node
```
elect launch node 1234 localhost:2345 localhost:3456
```

The above command launches a single subprocess listening on port 1234, and looking for peers running at localhost:2345 and localhost:3456, respectively. You can use this version of the command for deployments.
