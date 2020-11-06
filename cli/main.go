package main

import (
	"ethereum/keyservice/services/utils"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"sort"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app *cli.App

	KeyStoreFlag = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	KeyFlag = cli.StringFlag{
		Name:  "key",
		Usage: "Private key file path",
		Value: "",
	}
	RootFlag = cli.StringFlag{
		Name:  "root",
		Usage: "Root address",
		Value: "",
	}
	NameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "Dapp name",
		Value: "",
	}
	IpsFlag = cli.StringFlag{
		Name:  "ips",
		Usage: "Set Dapp ips, each separated , over",
		Value: "",
	}
	DescFlag = cli.StringFlag{
		Name:  "desc",
		Usage: "Dapp describe info",
		Value: "",
	}
	IDFlag = cli.StringFlag{
		Name:  "dappid",
		Usage: "Dapp id",
		Value: "",
	}
	CountFlag = cli.Uint64Flag{
		Name:  "count",
		Usage: "Derive account count",
		Value: 0,
	}
	StatusFlag = cli.Uint64Flag{
		Name:  "status",
		Usage: "Lock 0, Unlock 1,Default Lock",
		Value: 0,
	}
	AddressFlag = cli.StringFlag{
		Name:  "address",
		Usage: "Account address",
		Value: "",
	}
	RegisterFlags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		NameFlag,
		IpsFlag,
		DescFlag,
	}
	DeriveFlags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		IDFlag,
		IpsFlag,
		CountFlag,
	}
	UpdateDappFlags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		IDFlag,
		IpsFlag,
		DescFlag,
		StatusFlag,
	}
	UpdateAccountFlags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		IDFlag,
		IpsFlag,
		DescFlag,
		StatusFlag,
		AddressFlag,
	}
	DappAddressFlags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		IDFlag,
		AddressFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Usage = "TrueChain Impawn tool"
	app.Name = filepath.Base(os.Args[0])
	app.Version = "1.0.0"
	app.Copyright = "Copyright 2019-2020 The TrueChain Authors"
	app.Flags = []cli.Flag{
		KeyFlag,
		RootFlag,
		KeyStoreFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		IDFlag,
		NameFlag,
		IpsFlag,
		DescFlag,
		CountFlag,
		StatusFlag,
		AddressFlag,
	}
	app.CommandNotFound = func(ctx *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "No such command: %s\n", cmd)
		os.Exit(1)
	}
	// Add subcommands.
	app.Commands = []cli.Command{
		RegisterCommand,
		DappDeriveCommand,
		UpdateDappCommand,
		UpdateAccountCommand,
		DappAddressCommand,
	}
	cli.CommandHelpTemplate = utils.OriginCommandHelpTemplate
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
