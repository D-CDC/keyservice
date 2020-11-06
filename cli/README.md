## TrueChain KeyService CLI

TrueChain KeyService CLI is a tool, which can call `KeyService` participate in control account.

<a href="https://github.com/truechain/truechain-engineering-code/blob/master/COPYING"><img src="https://img.shields.io/badge/license-GPL%20%20truechain-lightgrey.svg"></a>

## Building the source


Building impawn requires both a Go (version 1.9 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    go build -o impawn  main.go dapp_admin.go admin_operate.go


### Command

The impawn project comes with several Sub Command.

|    SubCommand    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| :-----------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|  `register`   | Register dapp info.          |
|   `derive`    | Dapp Derive account. |
|  `updatedapp`   | Update dapp config.                                             |
|  `updateaccount`     | Update account config.                  |
| `dappaddress` | Dapp list a ccount or all account.              |
### Flag
  * `--key` Specify a file which contains private key as wallet seed. 
  * `--keystore` Specify a file which contains private key as wallet seed. 
  * `--root`       Root addres
  * `--rpcaddr` HTTP-RPC server listening interface (default: `localhost`)
  * `--rpcport` HTTP-RPC server listening port (default: `8545`)
  * `--dappid`    Dapp id
  * `--name`      Dapp name
  * `--ips`       Set Dapp ips, each separated , over
  * `--desc`      Dapp describe info
  * `--count`     Derive account count (default: 0)
  * `--status`    Lock 0, Unlock 1,Default Lock (default: 0)
  * `--address`   Account address
  
## Running CLI

### Register

```
$ ./main --keystore UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx --rpcaddr 39.100.97.xxx --rpcport 8985 --root "0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD" register --name "second dapp" --ips "127.0.0.1" --desc "my account2"

```

This command explain:
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx file.
 * `--rpcaddr` `--rpcport` flag show connect node ip + port.
 * `--root`       Root addres
 * **register**    sub command
 * `--name`      Dapp name
 * `--ips`       Set Dapp ips, each separated , over
 * `--desc`      Dapp describe info
  
**Output Log**
```shell
Connect url  http://localhost:8985  root  0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD  admin  0x7C017928A3DD264CB866aAA8a11cda0AAd4490A3
truekey_registerDapp [ID:0x5da7fd42ce37bd394cde3cc6014d0f5c27f90744578b6653361def6d5ce9d4d1 Note: Pub:0x047d0074afb1160ad855c0f721c5af951e14455e3bf29ee21d252014493d6b000825dc91890e5597bfe21f56d56e583a8fbb7c3b55f1937da28af7425408b9c11c]
```

### Derive

```
$ ./main --keystore UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx --rpcaddr 39.100.97.xxx --rpcport 8985 --root "0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD" derive  --ips "127.0.0.1"  --count 3 --dappid "0x3cc8f26e59895bf80be0668e51ba876f484adcb2dda7a9afb59aaf8c9de167ad"

```

This command explain:
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx file.
 * `--rpcaddr` `--rpcport` flag show connect node ip + port.
  * `--root`       Root addres
  * **derive**    sub command
  * `--dappid`    Dapp id
  * `--ips`       Set Dapp ips, each separated , over
  * `--count`     Derive account count (default: 0)

**Output Log**
```shell
Connect url  http://localhost:8985  root  0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD  admin  0x7C017928A3DD264CB866aAA8a11cda0AAd4490A3
truekey_dappDerive [[ID:36 Address:0x279fc1061D6e6Dc8942a73dfc4327FA2B31C1CE1 Status:1 
] [ID:37 Address:0xC00961825566af54cE847430bFa3feDfc81D3B77 Status:1 
] [ID:38 Address:0x620b0A490C3d3245EaD141550C6f6230Bb3502a9 Status:1 
]]
```

### UpdateDapp

```
$ ./main --keystore UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx --rpcaddr 39.100.97.xxx --rpcport 8985 --root "0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD" updatedapp --dappid 0x5da7fd42ce37bd394cde3cc6014d0f5c27f90744578b6653361def6d5ce9d4d1 --ips "127..0.0.1,127.0.0.2" --desc "my account2 update" --status 1

```

This command explain:
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx file.
 * `--rpcaddr` `--rpcport` flag show connect node ip + port.
 * `--root`       Root addres
 * **updatedapp**    sub command
 * `--dappid`    Dapp id
 * `--name`      Dapp name
 * `--ips`       Set Dapp ips, each separated , over
 * `--desc`      Dapp describe info
 * `--status`    Lock 0, Unlock 1,Default Lock (default: 0)

  
**Output Log**
```shell
Connect url  http://localhost:8985  root  0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD  admin  0x7C017928A3DD264CB866aAA8a11cda0AAd4490A3
updateDappp Success
```

### UpdateAccount

```
$ /main --keystore UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx --rpcaddr 39.100.97.xxx --rpcport 8985 --root "0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD" updateaccount   --dappid "0x3cc8f26e59895bf80be0668e51ba876f484adcb2dda7a9afb59aaf8c9de167ad" --address 0x279fc1061D6e6Dc8942a73dfc4327FA2B31C1CE1 --status 1 --ips "*.*.*.*" --desc "my account"


```

This command explain:
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx file.
 * `--rpcaddr` `--rpcport` flag show connect node ip + port.
  * `--root`       Root addres
  * **updateaccount**    sub command
  * `--dappid`    Dapp id
  * `--address`   Account address
  * `--ips`       Set Dapp ips, each separated , over
  * `--desc`      Dapp describe info
  * `--status`    Lock 0, Unlock 1,Default Lock (default: 0)

**Output Log**
```shell
Connect url  http://localhost:8985  root  0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD  admin  0x7C017928A3DD264CB866aAA8a11cda0AAd4490A3
updateAccount Success
```

### DappAddress

```
$ ./main --keystore UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx --rpcaddr 39.100.97.xxx --rpcport 8985 --root "0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD" dappaddress   --dappid "0x3cc8f26e59895bf80be0668e51ba876f484adcb2dda7a9afb59aaf8c9de167ad" --address 0x279fc1061D6e6Dc8942a73dfc4327FA2B31C1CE1

```

This command explain:
 * `--keystore` flag show load private key in UTC--2018-09-07T07-45-16.954721700Z--xxxxxxxxxx file.
 * `--rpcaddr` `--rpcport` flag show connect node ip + port.
  * `--root`       Root addres
  * **dappaddress**    sub command
  * `--dappid`    Dapp id **if not specify --dappid will list all dapps**
  * `--address`   Account address **if not specify --address will list all address in dapp**

**Output Log**
```shell
Connect url  http://localhost:8985  root  0x0EB4d5C43e894B42aaE58D859Cf926afA6A846BD  admin  0xfb65DAD4A88AfE5d45BD2c867AAC3a3f599265d8
truekey dappAddress
[ ID:0x6a6fc3ca8e0cbfc8a8dfbd0a82453a5a1c603d1ca1e32071c50a0add51a68826 Status: 1 Desc: my account1 update Ips: [127.0.0.1,127.0.0.2]  priv :0x6b074c690328747cbe1ad408b9318d40cc808884ef3bd059f1d25f2af753e863 ]
 Accounts [Account ID:0x937C6815B0b78C403beebf662C93dAf8A6111020 Status:1 Ips:[127.0.0.1] ][Account ID:0xb67b928F2a2C4DfB58c250f9525b7aAbf520173a Status:1 Ips:[127.0.0.1] ][Account ID:0xe3fBF0441368765AAE21562279dE249c8100f926 Status:1 Ips:[*.*.*.*] ]
[ ID:0x959438691b51888ba52d16cc6e318bd32ebd1174089afbf12e71ea8a1615eb08 Status: 1 Desc: my account2 Ips: [127.0.0.1]  priv :0x98f73e1a26879c5a9b4f9fc6037f6f967382f196702a9d960ba1b5f5634b8372 ]
 Accounts [Account ID:0xd19BaC914Cf882afce5307890BAc0cD54C8507a2 Status:1 Ips:[127.0.0.1] ][Account ID:0x1f19244d9bB11edfb5599A79df2fF3fDCc0EEe09 Status:1 Ips:[127.0.0.1] ][Account ID:0x1Cb6907756A11444412dB3B02034f8684aDaA2A7 Status:1 Ips:[127.0.0.1] ]
```