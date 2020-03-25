## TesraMainChain Go

Branch    | Tests 
----------|-------
master    | [![CircleCI](https://circleci.com/gh/TesraSupernet/TesraMainChain/tree/master.svg?style=shield)](https://circleci.com/gh/TesraSupernet/TesraMainChain/tree/master) 
develop   | [![CircleCI](https://circleci.com/gh/TesraSupernet/TesraMainChain/tree/develop.svg?style=shield)](https://circleci.com/gh/TesraSupernet/TesraMainChain/tree/develop) 


## Building the source

Building tesramain requires both a Go (version 1.7 or later) and a C compiler.

If build release version,Docker is required

You can install them using your favourite package manager.
Once the dependencies are installed, run

    make tesramain

or, to build the full suite of utilities:

    make all
	
or, to build the release version

    make release	

## Running tesramain

### Full node on the main tesramainchain network

By far the most common scenario is people wanting to simply interact with the tesramainchain network:
create accounts; transfer funds; deploy and interact with contracts. For this particular use-case
the user doesn't care about years-old historical data, so we can fast-sync quickly to the current
state of the network. To do so:

```
$ tesramain console
```

This command will:

 * Start tesramain in fast sync mode (default, can be changed with the `--syncmode` flag), causing it to
   download more data in exchange for avoiding processing the entire history of the tesramainchain network,
   which is very CPU intensive.
   This too is optional and if you leave it out you can always attach to an already running tesramain instance
   with `tesramain attach`.

### Full node on the tesramainchain test network

Transitioning towards developers, if you'd like to play around with creating tesramainchain contracts, you
almost certainly would like to do that without any real money involved until you get the hang of the
entire system. In other words, instead of attaching to the main network, you want to join the **test**
network with your node, which is fully equivalent to the main network, but with play-Ether only.

```
$ tesramain --testnet console
```

The `console` subcommand have the exact same meaning as above and they are equally useful on the
testnet too. Please see above for their explanations if you've skipped to here.

Specifying the `--testnet` flag however will reconfigure your tesramain instance a bit:

 * Instead of using the default data directory (`~/.tesramainchain` on Linux for example), tesramain will nest
   itself one level deeper into a `testnet` subfolder (`~/.tesramainchain/testnet` on Linux). Note, on OSX
   and Linux this also means that attaching to a running testnet node requires the use of a custom
   endpoint since `tesramain attach` will try to attach to a production node endpoint by default. E.g.
   `tesramain attach <datadir>/testnet/tesramain.ipc`. Windows users are not affected by this.
 * Instead of connecting the main tesramainchain network, the client will connect to the test network,
   which uses different P2P bootnodes, different network IDs and genesis states.
   
*Note: Although there are some internal protective measures to prevent transactions from crossing
over between the main network and test network, you should make sure to always use separate accounts
for play-money and real-money. Unless you manually move accounts, tesramain will by default correctly
separate the two networks and will not make any accounts available between them.*

### Programatically interfacing tesramain nodes

As a developer, sooner rather than later you'll want to start interacting with tesramain and the tesramainchain
network via your own programs and not manually through the console. To aid this, tesramain has built in
support for a JSON-RPC based APIs 。These can be
exposed via HTTP, WebSockets and IPC (unix sockets on unix based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by tesramain, whereas the HTTP
and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons.
These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr` HTTP-RPC server listening interface (default: "localhost")
  * `--rpcport` HTTP-RPC server listening port (default: 8545)
  * `--rpcapi` API's offered over the HTTP-RPC interface (default: "eth,net,web3")
  * `--rpccorsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr` WS-RPC server listening interface (default: "localhost")
  * `--wsport` WS-RPC server listening port (default: 8546)
  * `--wsapi` API's offered over the WS-RPC interface (default: "eth,net,web3")
  * `--wsorigins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: "admin,debug,eth,miner,net,personal,shh,txpool,web3")
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to connect
via HTTP, WS or IPC to a tesramain node configured with the above flags and you'll need to speak [JSON-RPC](http://www.jsonrpc.org/specification)
on all transports. You can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based transport before
doing so! Hackers on the internet are actively trying to subvert tesramainchain nodes with exposed APIs!
Further, all browser tabs can access locally running webservers, so malicious webpages could try to
subvert locally available APIs!**

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to start a
bootstrap node that others can use to find each other in your network and/or over the internet. The
clean way is to configure and run a dedicated bootnode:

```
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an [`enode` URL]
that other nodes can use to connect to it and exchange peer information. Make sure to replace the
displayed IP address information (most probably `[::]`) with your externally accessible IP to get the
actual `enode` URL.

*Note: You could also use a full fledged tesramain node as a bootnode, but it's the less recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try `telnet <ip> <port>` to ensure
it's indeed reachable), start every subsequent tesramain node pointed to the bootnode for peer discovery
via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your
private network separated, so do also specify a custom `--datadir` flag.

```
$ tesramain --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll also
need to configure a miner to process transactions and create new blocks for you.*

#### Docker quick start

One of the quickest ways to get tesramainchain up and running on your machine is by using Docker:

```
docker run -d --name tesramainchain-node -v /home/ubuntu/TesraMainChain:/root \
           -p 8545:8545 -p 17717:17717 \
           tesramainchain/client-go --rpc
```

This will start tesramain in fast-sync mode with a DB memory allowance of 1GB just as the above command does.  It will also create a persistent volume in your home directory for saving your blockchain as well as map the default ports. 
Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `tesramain` binds to the local interface and RPC endpoints is not accessible from the outside.

