## KeyService

WARNING!

TrueKeyService is an account management tool. It may, like any software, contain bugs.

Please take care to
- backup your keystore files,
- verify that the keystore(s) can be opened with your password.

TrueKeyService is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
PURPOSE. See the GNU General Public License for more details.

## Building the source


Building impawn requires both a Go (version 1.9 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

       go build services/truekey/main.go

### Defining the Admin Control CLI

First, you'll need to create config.json file, This consists of a small JSON file:

```json
{
    "rpcport": 8985,
    "rpcaddr": "127.0.0.1",
    "admins": [
        {
            "root": "0xc02f50f4f41f46b6a2f08036ae65039b2f9acd69"
        },
        {
            "root": "0x703c4b2bd70c169f5717101caee543299fc946c7"
        }
    ]
}
```

* `rpcport` Specify port for `CLI`
* `rpcaddr` Will listen all ip address for cli when giving `--rpcaddr 0.0.0.0`, you can give the exact ip address that want to connect, or `--rpcaddr 127.0.01` only allow running on the host to connect `service`.
* `root`    Specify root keystore address
* `admins`  Accept which `CLI` connections 

### Start Service

```
$ ./main --datadir data --config data/config.json  --keystore data/UTC--2020-09-10T08-42-10.662467000Z--e4fad2e5ee2e878e65f1fe02c0f9edaf54789a8e --rpcaddr "0.0.0.0" --rpc

```

This command explain:
 * `--datadir`  specify data dir for store. 
 * `--config`   specify admin config for cli.
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx for wallet seed.
  * `--keystoredir` flag show load private key in directory for wallet seed.
 * `--rpcaddr` `--rpcport` this for **dapp** connections,Will listen all ip address for cli when giving `--rpcaddr 0.0.0.0`, you can give the exact ip address that want to connect, or `--rpcaddr 127.0.01` only allow running on the host to connect `service`.
 * `--rpc`  enable rpc function.
