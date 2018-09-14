package main

import (
	"path/filepath"

	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrdata/v3/netparams"
)

const (
	defaultConfigFilename = "dcrdatawallet.conf"
	defaultLogFilename    = "dcrdatawallet.log"
	defaultDataDirname    = "data"
	defaultLogLevel       = "info"
	defaultLogDirname     = "logs"
)

var activeNet = &netparams.MainNetParams
var activeChain = &chaincfg.MainNetParams

var (
	defaultHomeDir           = dcrutil.AppDataDir("dcrdatawallet", false)
	defaultConfigFile        = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultLogDir            = filepath.Join(defaultHomeDir, defaultLogDirname)
	defaultDataDir           = filepath.Join(defaultHomeDir, defaultDataDirname)
	dcrdHomeDir              = dcrutil.AppDataDir("dcrd", false)
	defaultDaemonRPCCertFile = filepath.Join(dcrdHomeDir, "rpc.cert")

	defaultHost = "localhost"

	defaultPGHost   = "127.0.0.1:5432"
	defaultPGUser   = "dcrdata"
	defaultPGPass   = ""
	defaultPGDBName = "dcrdata"
	// default Account
	defaultWalletServer1 = "127.0.0.1:19101"
	defaultWalletCert1   = "/home/casey/.dcrwallet-cold-default/rpc.cert"
	// june-coins Account
	defaultWalletServer2 = "127.0.0.1:19102"
	defaultWalletCert2   = "/home/casey/.dcrwallet-cold-june-coins/rpc.cert"
	// staking-rewards
	defaultWalletServer3 = "127.0.0.1:19103"
	defaultWalletCert3   = "/home/casey/.dcrwallet-cold-staking-rewards/rpc.cert"
	// Hot Wallet (Voter)
	defaultWalletServer4 = "127.0.0.1:9110"
	defaultWalletCert4   = "/home/casey/.dcrwallet/rpc.cert"
)

type config struct {
	// General application behavior
	HomeDir     string `short:"A" long:"appdata" description:"Path to application home directory" env:"DCRDATA_APPDATA_DIR"`
	ConfigFile  string `short:"C" long:"configfile" description:"Path to configuration file" env:"DCRDATA_CONFIG_FILE"`
	DataDir     string `short:"b" long:"datadir" description:"Directory to store data" env:"DCRDATA_DATA_DIR"`
	LogDir      string `long:"logdir" description:"Directory to log output." env:"DCRDATA_LOG_DIR"`
	OutFolder   string `short:"f" long:"outfolder" description:"Folder for file outputs" env:"DCRDATA_OUT_FOLDER"`
	ShowVersion bool   `short:"V" long:"version" description:"Display version information and exit"`

	// RPC client options
	WalletServer1    string `long:"walletserver1" description:"Hostname/IP and port where dcrwallet (default account) is listening on"`
	WalletCert1      string `long:"walletcert1" description:"File containing the dcrwallet certificate for default account wallet"`
	WalletServer2    string `long:"walletserver2" description:"Hostname/IP and port where dcrwallet (june-coins account) is listening on"`
	WalletCert2      string `long:"walletcert2" description:"File containing the dcrwallet certificate for june-coins wallet"`
	WalletServer3    string `long:"walletserver3" description:"Hostname/IP and port where dcrwallet (staking-rewards account) is listening on"`
	WalletCert3      string `long:"walletcert3" description:"File containing the dcrwallet certificate for staking rewards wallet"`
	WalletServer4    string `long:"walletserver4" description:"Hostname/IP and port where dcrwallet (hot voter account) is listening on"`
	WalletCert4      string `long:"walletcert4" description:"File containing the dcrwallet certificate for hot wallet default account"`
	DcrdUser         string `long:"dcrduser" description:"Daemon RPC user name" env:"DCRDATA_DCRD_USER"`
	DcrdPass         string `long:"dcrdpass" description:"Daemon RPC password" env:"DCRDATA_DCRD_PASS"`
	DcrdServ         string `long:"dcrdserv" description:"Hostname/IP and port of dcrd RPC server to connect to (default localhost:9109, testnet: localhost:19109, simnet: localhost:19556)" env:"DCRDATA_DCRD_URL"`
	DcrdCert         string `long:"dcrdcert" description:"File containing the dcrd certificate file" env:"DCRDATA_DCRD_CERT"`
	DisableDaemonTLS bool   `long:"nodaemontls" description:"Disable TLS for the daemon RPC client -- NOTE: This is only allowed if the RPC client is connecting to localhost" env:"DCRDATA_DCRD_DISABLE_TLS"`
}

var (
	defaultConfig = config{
		HomeDir:       defaultHomeDir,
		DataDir:       defaultDataDir,
		LogDir:        defaultLogDir,
		ConfigFile:    defaultConfigFile,
		DcrdCert:      defaultDaemonRPCCertFile,
		WalletServer1: defaultWalletServer1,
		WalletServer2: defaultWalletServer2,
		WalletServer3: defaultWalletServer3,
		WalletServer4: defaultWalletServer4,
		WalletCert1:   defaultWalletCert1,
		WalletCert2:   defaultWalletCert2,
		WalletCert3:   defaultWalletCert3,
		WalletCert4:   defaultWalletCert4,
	}
)

// loadConfig initializes and parses the config using a config file and command
// line options.
func loadConfig() (*config, error) {
	loadConfigError := func(err error) (*config, error) {
		return nil, err
	}
	// Default config.
	cfg := defaultConfig

	// Find the active network and latch onto it
	activeNet = &netparams.MainNetParams
	activeChain = &chaincfg.MainNetParams

	// Set the host names and ports to the default if the user does not specify
	// them.
	if cfg.DcrdServ == "" {
		cfg.DcrdServ = defaultHost + ":" + activeNet.JSONRPCClientPort
	}

	// Set the host name and port of the walletservers
	// if the user does not specify them
	if cfg.WalletServer1 == "" {
		cfg.WalletServer1 = defaultWalletServer1
	}
	if cfg.WalletServer2 == "" {
		cfg.WalletServer2 = defaultWalletServer2
	}
	if cfg.WalletServer3 == "" {
		cfg.WalletServer3 = defaultWalletServer3
	}
	if cfg.WalletServer4 == "" {
		cfg.WalletServer4 = defaultWalletServer4
	}
	if cfg.WalletCert1 == "" {
		cfg.WalletCert1 = defaultWalletCert1
	}
	if cfg.WalletCert2 == "" {
		cfg.WalletCert2 = defaultWalletCert2
	}
	if cfg.WalletCert3 == "" {
		cfg.WalletCert3 = defaultWalletCert3
	}
	if cfg.WalletCert4 == "" {
		cfg.WalletCert4 = defaultWalletCert4
	}

	return &cfg, nil
}
