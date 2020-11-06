// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"ethereum/keyservice/accounts"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/fdlimit"
	"ethereum/keyservice/etruedb"
	"ethereum/keyservice/services/truekey/signer"
	"ethereum/keyservice/services/truekey/types"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"ethereum/keyservice/accounts/keystore"
	"ethereum/keyservice/console"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rpc"
	"ethereum/keyservice/services/utils"
	colorable "github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"gopkg.in/urfave/cli.v1"
)

const legalWarning = `
WARNING!

TrueKeyService is an account management tool. It may, like any software, contain bugs.

Please take care to
- backup your keystore files,
- verify that the keystore(s) can be opened with your password.

TrueKeyService is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
PURPOSE. See the GNU General Public License for more details.
`

const (
	DatabaseCache   = 256
	KEYDataDir      = "keydata"
	ServerAUDITFILE = "server_audit.log"
	ConfigFile      = "config.json"
	DefaultHTTPHost = "localhost" // Default host interface for the HTTP RPC server
	DefaultHTTPPort = 8545        // Default TCP port for the HTTP RPC server
)

var (
	logLevelFlag = cli.IntFlag{
		Name:  "loglevel",
		Value: 4,
		Usage: "log level to emit to the screen",
	}
	keystoreFlag = cli.StringFlag{
		Name:  "keystore",
		Usage: "Keystore file path",
	}
	keystoreDirFlag = cli.StringFlag{
		Name:  "keystoredir",
		Usage: "Keystore directory path",
	}
	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Value: DefaultConfigDir(),
		Usage: "Directory for TrueKeyService configuration",
	}
	rpcPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: DefaultHTTPPort + 5,
	}
	ConfigFlag = cli.StringFlag{
		Name:  "config",
		Usage: "Config file path",
	}
	app         = cli.NewApp()
	initCommand = cli.Command{
		Action:    utils.MigrateFlags(initializeKeyStore),
		Name:      "init",
		Usage:     "Initialize the signer, generate a keystore",
		ArgsUsage: "",
		Flags: []cli.Flag{
			logLevelFlag,
			DataDirFlag,
		},
		Description: `
The init command generates a keystore which TrueKeyService can use to start service.`,
	}
)

func init() {
	app.Name = "TrueKeyService"
	app.Usage = "Manage TrueChain account operations"
	app.Flags = []cli.Flag{
		logLevelFlag,
		keystoreFlag,
		keystoreDirFlag,
		DataDirFlag,
		utils.LightKDFFlag,
		utils.RPCListenAddrFlag,
		utils.RPCVirtualHostsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
		utils.RPCEnabledFlag,
		rpcPortFlag,
		ConfigFlag,
	}
	app.Action = trueKeyService
	app.Commands = []cli.Command{initCommand}
	cli.CommandHelpTemplate = utils.OriginCommandHelpTemplate
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initializeKeyStore(c *cli.Context) error {
	// Get past the legal message
	if err := initialize(c); err != nil {
		return err
	}
	// Ensure the master key does not yet exist, we're not willing to overwrite
	configDir := c.GlobalString(DataDirFlag.Name)
	if err := os.Mkdir(configDir, 0700); err != nil && !os.IsExist(err) {
		return err
	}
	n, p := keystore.StandardScryptN, keystore.StandardScryptP
	if c.GlobalBool(utils.LightKDFFlag.Name) {
		n, p = keystore.LightScryptN, keystore.LightScryptP
	}
	text := "Please specify a password. Do not forget this password!"
	var password string
	for {
		password = getPassPhrase(text, true)
		if err := signer.ValidatePasswordFormat(password); err != nil {
			fmt.Printf("invalid password: %v\n", err)
		} else {
			fmt.Println()
			break
		}
	}

	ks := keystore.NewKeyStore(configDir, n, p)
	account, err := ks.NewAccount(password)
	if err != nil {
		return err
	}
	log.Info("Initialize keystore success", "keystore path", configDir, "address", account.Address.String())
	return nil
}

func initialize(c *cli.Context) error {
	// Set up the logger to print everything
	logOutput := os.Stdout
	if !confirm(legalWarning) {
		return fmt.Errorf("aborted by user")
	}
	fmt.Println()
	usecolor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(logOutput)
	if usecolor {
		output = colorable.NewColorable(logOutput)
	}
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(c.Int(logLevelFlag.Name)), log.StreamHandler(output, log.TerminalFormat(usecolor))))

	return nil
}

// ipcEndpoint resolves an IPC endpoint based on a configured value, taking into
// account the set data folders as well as the designated platform we're currently
// running on.
func ipcEndpoint(ipcPath, datadir string) string {
	// On windows we can only use plain top-level pipes
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(ipcPath, `\\.\pipe\`) {
			return ipcPath
		}
		return `\\.\pipe\` + ipcPath
	}
	// Resolve names into the data directory full paths otherwise
	if filepath.Base(ipcPath) == ipcPath {
		if datadir == "" {
			return filepath.Join(os.TempDir(), ipcPath)
		}
		return filepath.Join(datadir, ipcPath)
	}
	return ipcPath
}

func trueKeyService(c *cli.Context) error {
	// If we have some unrecognized command, bail out
	if args := c.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}
	if err := initialize(c); err != nil {
		return err
	}

	configDir := c.GlobalString(DataDirFlag.Name)
	stretchedKey, err := readMasterKey(c)
	if err != nil {
		fmt.Println("Please use init command init a keystore or specified correct keystore")
		return fmt.Errorf("aborted by user err = %s", err.Error())
	}
	keydata, err := etruedb.NewLDBDatabase(filepath.Join(configDir, KEYDataDir), DatabaseCache, makeDatabaseHandles())
	if err != nil {
		log.Info("NewLDBDatabase", "err", err)
		return err
	}
	var (
		rootLoc    = filepath.Join(configDir, "keystore")
		lightKdf   = c.GlobalBool(utils.LightKDFFlag.Name)
		configFile = c.GlobalString(ConfigFlag.Name)
	)

	if !c.GlobalIsSet(ConfigFlag.Name) {
		configFile = filepath.Join(configDir, "config.json")
	}

	configAdmins := types.LoadNodesJSON(configFile)
	log.Info("Starting signer", "keystore", rootLoc, "light-kdf", lightKdf)
	apiImpl, err := signer.NewSignerAPI(keydata, stretchedKey, configAdmins.Config)
	if err != nil {
		log.Info("NewSignerAPI", "err", err)
		return err
	}

	// Establish the bidirectional communication, by creating a new UI backend and registering
	// it with the UI.
	//ServerAUDITFILE
	truekeyApi, err := signer.NewServerAuditLogger(ServerAUDITFILE, signer.NewUIServerAPI(apiImpl))
	if err != nil {
		utils.Fatalf(err.Error())
	}
	log.Info("Audit server logs configured", "file", ServerAUDITFILE)

	// register signer API with server
	var (
		extapiURL = "n/a"
		ipcapiURL = "n/a"
	)

	serverAPI := []rpc.API{
		{
			Namespace: "truekey",
			Public:    true,
			Service:   truekeyApi,
			Version:   "1.0"},
	}

	vhosts := splitAndTrim(c.GlobalString(utils.RPCVirtualHostsFlag.Name))
	cors := splitAndTrim(c.GlobalString(utils.RPCCORSDomainFlag.Name))
	// start http server
	httpEndpoint := fmt.Sprintf("%s:%d", c.GlobalString(utils.RPCListenAddrFlag.Name), c.Int(rpcPortFlag.Name))
	listener, _, err := rpc.StartHTTPEndpoint(httpEndpoint, serverAPI, []string{"truekey"}, cors, vhosts)
	if err != nil {
		utils.Fatalf("Could not start RPC api: %v", err)
	}
	extapiURL = fmt.Sprintf("http://%s", httpEndpoint)
	log.Info("HTTP endpoint server opened", "url", extapiURL)

	defer func() {
		listener.Close()
		log.Info("HTTP endpoint server closed", "url", httpEndpoint)
	}()

	if !c.GlobalBool(utils.IPCDisabledFlag.Name) {
		givenPath := c.GlobalString(utils.IPCPathFlag.Name)
		ipcapiURL = ipcEndpoint(filepath.Join(givenPath, "truekey.ipc"), configDir)
		listener, _, err := rpc.StartIPCEndpoint(ipcapiURL, serverAPI)
		if err != nil {
			utils.Fatalf("Could not start IPC api: %v", err)
		}
		log.Info("IPC endpoint opened", "url", ipcapiURL)
		defer func() {
			listener.Close()
			log.Info("IPC endpoint closed", "url", ipcapiURL)
		}()
	}

	abortChan := make(chan os.Signal, 1)
	signal.Notify(abortChan, os.Interrupt)

	sig := <-abortChan
	apiImpl.Stop()
	log.Info("Exiting...", "signal", sig)

	return nil
}

// splitAndTrim splits input separated by a comma
// and trims excessive white space from the substrings.
func splitAndTrim(input string) []string {
	result := strings.Split(input, ",")
	for i, r := range result {
		result[i] = strings.TrimSpace(r)
	}
	return result
}

// DefaultConfigDir is the default config directory to use for the vaults and other
// persistence requirements.
func DefaultConfigDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Signer")
		} else if runtime.GOOS == "windows" {
			appdata := os.Getenv("APPDATA")
			if appdata != "" {
				return filepath.Join(appdata, "Signer")
			} else {
				return filepath.Join(home, "AppData", "Roaming", "Signer")
			}
		} else {
			return filepath.Join(home, ".truekey")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func printError(error ...interface{}) {
	log.Crit("Error", "err", error)
}

// KeyStoreScheme is the protocol scheme prefixing account and wallet URLs.
const KeyStoreScheme = "keystore"

func getAllKeystoreFile(path string) ([]string, error) {
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		printError("path ", err)
		return nil, err
	}
	var files []string
	for _, fi := range rd {
		if fi.IsDir() {
			continue
		} else {
			var (
				buf = new(bufio.Reader)
				key struct {
					Address string `json:"address"`
				}
			)
			readAccount := func(path string) *accounts.Account {
				fd, err := os.Open(path)
				if err != nil {
					log.Trace("Failed to open keystore file", "path", path, "err", err)
					return nil
				}
				defer fd.Close()
				buf.Reset(fd)
				// Parse the address.
				key.Address = ""
				err = json.NewDecoder(buf).Decode(&key)
				addr := common.HexToAddress(key.Address)
				switch {
				case err != nil:
				case (addr == common.Address{}):
				default:
					return &accounts.Account{Address: addr, URL: accounts.URL{Scheme: KeyStoreScheme, Path: path}}
				}
				return nil
			}

			if a := readAccount(filepath.Join(path, fi.Name())); a != nil {
				files = append(files, filepath.Join(path, fi.Name()))
			}
		}
	}
	return files, nil
}

func readMasterKey(ctx *cli.Context) ([]*keystore.Key, error) {
	var (
		files []string
		keys  []*keystore.Key
	)
	find := false
	if ctx.GlobalIsSet(keystoreDirFlag.Name) {
		find = true
		keystores, err := getAllKeystoreFile(ctx.GlobalString(keystoreDirFlag.Name))
		if err != nil {
			return nil, err
		}
		if len(keystores) == 0 {
			return nil, errors.New("no find keystore, please specified correct dir")
		}
		files = append(files, keystores...)
	}

	if ctx.GlobalIsSet(keystoreFlag.Name) {
		find = true
		files = append(files, ctx.GlobalString(keystoreFlag.Name))
	}

	if !find {
		return nil, errors.New("please specified keystore")
	}

	for _, keyfile := range files {
		keyjson, err := ioutil.ReadFile(keyfile)
		if err != nil {
			return nil, fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err)
		}
		for trials := 0; trials < 2; trials++ {
			password := getPassPhrase("Please enter the password to decrypt the'"+keyfile+"':", false)

			//password := "secret"
			key, err := keystore.DecryptKey(keyjson, password)
			if err != nil {
				if trials == 1 {
					fmt.Println("Input error password limit ", keyfile)
					break
				}
				fmt.Println("Please input correct password")
			} else {
				log.Info("Decrypt keystore success", "keyfile", keyfile)
				keys = append(keys, key)
				break
			}
		}
	}

	if len(keys) == 0 {
		return keys, errors.New("no keystore can use")
	}

	return keys, nil
}

// confirm displays a text and asks for user confirmation
func confirm(text string) bool {
	fmt.Print(text)
	return true
}

// getPassPhrase retrieves the password associated with truekey, either fetched
// from a list of preloaded passphrases, or requested interactively from the user.
func getPassPhrase(prompt string, confirmation bool) string {
	fmt.Println(prompt)
	password, err := console.Stdin.PromptPassword("Password: ")
	if err != nil {
		utils.Fatalf("Failed to read password: %v", err)
	}
	if confirmation {
		confirm, err := console.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			utils.Fatalf("Failed to read password confirmation: %v", err)
		}
		if password != confirm {
			utils.Fatalf("Passwords do not match")
		}
	}
	return password
}

type encryptedSeedStorage struct {
	Description string              `json:"description"`
	Version     int                 `json:"version"`
	Params      keystore.CryptoJSON `json:"params"`
}

// encryptSeed uses a similar scheme as the keystore uses, but with a different wrapping,
// to encrypt the master seed
func encryptSeed(seed []byte, auth []byte, scryptN, scryptP int) ([]byte, error) {
	cryptoStruct, err := keystore.EncryptDataV3(seed, auth, scryptN, scryptP)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&encryptedSeedStorage{"TrueKeyService seed", 1, cryptoStruct})
}

// decryptSeed decrypts the master seed
func decryptSeed(keyjson []byte, auth string) ([]byte, error) {
	var encSeed encryptedSeedStorage
	if err := json.Unmarshal(keyjson, &encSeed); err != nil {
		return nil, err
	}
	if encSeed.Version != 1 {
		log.Warn(fmt.Sprintf("unsupported encryption format of seed: %d, operation will likely fail", encSeed.Version))
	}
	seed, err := keystore.DecryptDataV3(encSeed.Params, auth)
	if err != nil {
		return nil, err
	}
	return seed, err
}

// makeDatabaseHandles raises out the number of allowed file handles per process
// for Getrue and returns half of the allowance to assign to the database.
func makeDatabaseHandles() int {
	limit, err := fdlimit.Maximum()
	if err != nil {
		utils.Fatalf("Failed to retrieve file descriptor allowance: %v", err)
	}
	raised, err := fdlimit.Raise(uint64(limit))
	if err != nil {
		utils.Fatalf("Failed to raise file descriptor allowance: %v", err)
	}
	return int(raised / 2) // Leave half for networking and other stuff
}
