# Dev Notes: Integration with `HoneyBadgerMPC`
1. [Getting Started](#getting-started)
2. [Running the Unit Tests](#running-the-unit-tests)
3. [Running the Integration Tests](#running-the-integration-tests)
4. [Implementation](#implementation)
5. [End-to-end Demo](#end-to-end-demo)
6. [Continuous Integration: Travis](#continuous-integration-travis)

The goal of this document is provide instructions on how to get setup to perform
development work for the integration of HoneyBadgerMPC within Fabric.

A successful setup should allow one to:

* execute unit tests
* execute integration tests

Moreover, one should be capable to run specific tests, as it will probably be needed
during development.

Lastly, one should be capable to confirm that `honeybadgermpc` code is indeed available
and being executed when running specific tests.

Thus, one should be in a position to implement what is needed for the integration, along
with the necessary unit and integration tests.

This documentation also wishes to provide pointers to strategic parts of the fabric code
that are most relevant to the integration.


## Getting Started

**Get the prerequisites**

* `git`
* `go >= 1.10`
* `docker`
* `docker-compose`
* `python 3.7`

See https://hyperledger-fabric.readthedocs.io/en/latest/dev-setup/devenv.html#prerequisites
for more details.

**Set the GOPATH env var**

See https://github.com/golang/go/wiki/SettingGOPATH and
https://github.com/golang/go/wiki/GOPATH

**Clone the fabric code**

```bash
mkdir -p $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone -b hyperbadgermpc git@github.com:sbellem/fabric.git
cd fabric
```


## Running the Unit Tests
To run the tests, a setup phase is needed. This can be done via the `Makefile`:

```bash
make peer-docker testenv ccenv
```

### Running the "dummy" test with `honeybadgermpc.polynomial`

```bash
make hbmpc-tests
```

the above is the equivalent of:

```bash
cd unit-test
docker-compose run --rm --no-deps unit-tests \
    go test -v \
    github.com/hyperledger/fabric/core/handlers/endorsement/builtin/... \
    -run=TestDefaultEndorsement
```

There's also an additional target to run the tests, which will also re-build the docker
images (this takes time):

```bash
make hbmpc-tests-with-deps
```

### Run all tests
Must be in project root (not in `unit-test` dir).

```bash
cd $GOPATH/src/github.com/hyperledger/fabric 
make unit-tests
```

To run in verbose mode (i.e.: `go test -v ...`):

```bash
VERBOSE=1 make unit-tests
```

**Note** that running `make unit-tests` will most likely re-build some docker images and
will be time-consuming.

The `unit-test` target is defined as follows:

```makefile
unit-test: unit-test-clean peer-docker testenv ccenv
	cd unit-test && \
        docker-compose up --abort-on-container-exit --force-recreate && \
        docker-compose down
```

So if one does not need to run the docker related targets and etc, then one may simply
execute:

```bash
cd unit-test && \
    docker-compose up --abort-on-container-exit --force-recreate && \
    docker-compose down
```

Or, perhaps even more explicitly:

```bash
cd unit-test
docker-compose run --rm unit-tests ./unit-test/run.sh
```

This should save some time, if one needs to run the tests, e.g. after some code edits,
multiple times.

## Running the Integration Tests
Make sure you have [`Go >= 1.10`](https://golang.org/dl/).

Since the integration tests are not run in a docker container `honeybadgermpc` must
be available locally.

The preferred approach is to use a virtual environment, e.g.:

```bash
python3.7 -m venv hbmpc
source hbmpc/bin/activate
pip install -e /path/to/honeybadgermpc/code
# e.g:
# pip install -e ~/code/initc3/HoneyBadgerMPC/
```

Run all the integration tests:

```bash
make integration-test
```

To run a specific suite of tests, e.g.: `e2e`


```bash
./scripts/run-integration-tests.sh integration/e2e
```

or, if more control is needed the `ginkgo` command can be used directly:


if you have not ran `make integration-test`, you need the prerequisites:

```bash
make gotool.ginkgo ccenv docker-thirdparty
```

then, using the `ginkgo` command, you can run the tests for `e2e`:

```bash
ginkgo -keepGoing --slowSpecThreshold 60 -r integration/e2e/
```

for more fine grained control the `--focus` and `--skip` options can be used, or
modifications can be made to the code to focus on specific specs. See `ginkgo` docs
at https://onsi.github.io/ginkgo/#focused-specs to see how that works.

**Is the `honeybadgermpc` code executed?** Currently, there is a "dummy" call made
to `honeybadgermpc.polynomial`. To see its output when running the integration tests,
you have to run in verbose mode:

```bash
ginkgo -v -keepGoing --slowSpecThreshold 60 -r integration/e2e/

# ...

[o][Org1.peer1] combined out:
[o][Org1.peer1] {32772421984453869049654837817616228648556595312829773639127286687461613240324} + {26897919639731461304800066697180379259483881633235653031056949069575173982992} x^1 + {1310$
968793781547622893936849860937238351873964483684386637385585154337013760} x^2 + {28758063787632374594519486007194045141645120096579383362339064652588822332981} x^3 + {1966345319067232142979$
902690569737189133957187697864183476372012476967944190} x^4 + {12428986741613181552186880064403855879701830594144268252193105603804740433137} x^5 + {1310896879378154761682993340423204568049$
402285780134524664443764814953578496} x^6 + {10568842593712268267665177993500668475704996426832152088316368723939135027660} x^7
[o][Org1.peer1] omega1: {52435875175126190479447740508185965837690552500527637822603658699938581184512}
[o][Org1.peer1] omega2: {52435875175126190479447740508185965837690552500527637822603658699938581184512}
[o][Org1.peer1] eval:

# ...
```


## Implementation

_work in progress_

The clearest idea right now is to add an HBMPC system chaincode in the peer. A different
approach could be to use some kind of plugin, although this is not clear on how to do
so. Yet another idea would be to have HBMPC running as a "third-party" component in a
docker container.

One possible advantage of the plugin or the docker container is that it would limit the
need to build the peer image each time we make modifications to the HBMPC-related code.

### Add new system chaincode in peer
1. Implement a new system chaincode under [core/scc](./core/scc), say `hbmpcscc`.
2. Register that new system chaincode in [peer/node/start.go](./peer/node/start.go),
   under the function `registerChaincodeSupport`:

   ```go
   import (
           // ...
           "github.com/hyperledger/fabric/core/scc/cscc"
           "github.com/hyperledger/fabric/core/scc/lscc"
           "github.com/hyperledger/fabric/core/scc/qscc"

           // HoneyBadgerMPC system chaincode import
           "github.com/hyperledger/fabric/core/scc/hbmpcscc"
           // ...
   )
   // ...
   func registerChaincodeSupport(...) (...) {
          // ...

          lsccInst := lscc.New(sccp, aclProvider, pr)

          // ...

          csccInst := cscc.New(ccp, sccp, aclProvider)
          qsccInst := qscc.New(aclProvider)

          // HoneyBadgerMPC system chaincode instantiation
          hbmpcsccInst := hbmpcscc.New(...)

          //Now that chaincode is initialized, register all system chaincodes.
          sccs := scc.CreatePluginSysCCs(sccp)

          // HoneyBadgerMPC system chaincode registration along with others ...
          for _, cc := range append([]scc.SelfDescribingSysCC{lsccInst, csccInst, qsccInst, hbmpcsccInst, lifecycleSCC}, sccs...) {
                  sccp.RegisterSysCC(cc)
          }
          pb.RegisterChaincodeSupportServer(grpcServer.Server(), ccSrv)
          return chaincodeSupport, ccp, sccp
   }
   ```

### Plugin approach (?)
_to document_


### HBMPC as an independent containerized service (?)
_to document_


### Peer-to-peer Communication
New port mappings can be added in the `core.yaml` configuration.


## End-to-end Demo

_work in progress_

### end-to-end (e2e) examples
Look under [examples](./examples). There is some documentation under
[examples/e2e_cli/end-to-end.rst](examples/e2e_cli/end-to-end.rst).

In order to test and demo the integration we can copy and modify one of the examples.

### `fabric-samples`
The idea is to have a custom `Dockerfile` such that we can run the fabric code that has
the HoneyBadgerMPC integration.

As a first task, we need to identify which example (e.g. `first-network`) is most
relavant to the integration work.

To show how this could work, the `first-network` peer image has been changed so
that it is built using a local `Dockerfile` that [pulls the fabric code](
https://github.com/sbellem/fabric-samples/blob/hyperbadgermpc/first-network/base/Dockerfile#L76)
from a modified version of fabric.

Future work could be done so that the fabric code is obtained locally in order to allow
quick iterations. This could be done by having a git submodule or subtree for the fabric
code under the fabric-samples repository for instance.


## Continuous Integration: Travis
Some work has been to done have the tests run on travis. The reason for this is to
make it easier for collaboration.

Three jobs were added for `fabric`'s code (https://travis-ci.org/sbellem/fabric):

1. unit tests
2. `honeybadgermpc` related unit tests
3. integration tests

The `fabric-samples` tests were also put onto travis
(https://travis-ci.org/search/fabric-samples).

It should be noted that although the unit tests were successfully run there has been a
problem lately, and it remains to be fixed. The cause is not clear. Perhaps some changes
made to the travis code caused this, but this is just a blind guess.

Also, the integration tests have yet to be successfully run on travis. The cause of the
problem may some environment variable that is not properly set. But this is just a very
rough guess.
