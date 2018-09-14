// Copyright (c) 2018, Casey Friday

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"context"

	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrdata/v3/api"
	m "github.com/decred/dcrdata/v3/middleware"
	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

type appContext struct {
	coolDefault		*rpcclient.Client
	coolJuneCoins		*rpcclient.Client
	coolStakingRewards	*rpcclient.Client
	hotWallet		*rpcclient.Client
	nodeClient		*rpcclient.Client
	queryType		QueryType
}

type QueryType string
const (
	CoolDefault QueryType = "cool-default"	
	CoolJuneCoins QueryType = "cool-june-coins"
	CoolStakingRewards QueryType = "cool-staking-rewards"
	HotWallet QueryType = "hot-wallet"
	AllAccounts QueryType = "all"
	ctxAccount
}

func NewContext( wallet1 *rpcclient.Client, wallet2 *rpcclient.Client, wallet3 *rpcclient.Client, wallet4 *rpcclient.Client, node *rpcclient.Client, ) *appContext {
	return &appContext{
		coolDefault:	wallet1,
		coolJuneCoins:	wallet2,
		coolStakingRewards: wallet3,
		hotWallet:	wallet4,
		nodeClient:	node,
	}
}
// queryType is going to be 'tickets', 'transactions', etc

// mainCore does all the work. Deferred functions do not run after os.Exit(),
// so main wraps this function, which returns a code.
func main() {
	// Parse the configuration file.
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("Failed to load dcrdata config: %s\n", err.Error())
		return
	}

	// Start with version info
	//	log.Info("Dcrdata Wallet Tracking v1.0")
	var dcrdCerts []byte
	dcrdCerts, err = ioutil.ReadFile(cfg.DcrdCert)
	if err != nil {
		log.Printf("Failed to read dcrd cert at %s: %s\n", cfg.DcrdCert, err.Error())
	}
	// Connect to dcrd RPC server using websockets
	dcrdCfg := &rpcclient.ConnConfig{
		Host:         "localhost:9109",
		Endpoint:     "ws",
		User:         cfg.DcrdUser,
		Pass:         cfg.DcrdPass,
		Certificates: dcrdCerts,
	}
	// Daemon client connection

	dcrdClient, err := rpcclient.New(dcrdCfg, nil)
	if err != nil || dcrdClient == nil {
		return
	}

	// Initialize the dcrwallet funcs for accessing the dcrwallet daemons. This creates its own RPC
	//  connections since the dcrwallet daemons listen on a different port from the dcrd RPC server
	dcrwalletClient1 := WalletConnect(cfg.WalletServer1, cfg.DcrdUser, cfg.DcrdPass, cfg.WalletCert1, cfg.DisableDaemonTLS)
	dcrwalletClient2 := WalletConnect(cfg.WalletServer2, cfg.DcrdUser, cfg.DcrdPass, cfg.WalletCert2, cfg.DisableDaemonTLS)
	dcrwalletClient3 := WalletConnect(cfg.WalletServer3, cfg.DcrdUser, cfg.DcrdPass, cfg.WalletCert3, cfg.DisableDaemonTLS)
	dcrwalletClient4 := WalletConnect(cfg.WalletServer4, cfg.DcrdUser, cfg.DcrdPass, cfg.WalletCert4, cfg.DisableDaemonTLS)

	// Create a context that contains all wallets and the dcrd chain
	// Easier to use anything we need now, with this struct
	app, err := NewContext(dcrwalletClient1, dcrwalletClient2, dcrwalletClient3, dcrwalletClient4, dcrdClient)
	if err != nil {
		log.Printf("Could not create appContext type struct from all wallets and dcrd node")
	}

	// Create my own router to route traffic along the URL's I decide
	mux := chi.NewRouter()
	corsMW := cors.Default()
	mux.Use(corsMW.Handler)
	mux.Get("/", app.root) // TODO: write func
	mux.Route("/tx", func(r chi.Router) {
		r.Route("/", func(rt chi.Router) {
			rt.Route("/{account}", func(rd chi.Router) {
				rd.Use(AccountCtx)
				rd.Get("/", app.txGetter)
			})
		})
	})
	mux.Get("/tickets", casey.ticketgetter)

	mux.NotFound(func(w http.ResponseWriter, r *http.Request)) {
		http.Error(w, r.URL.RequestURI()+" does not seem to exist! (404)", http.StatusNotFound)
	})
	
	//mux.With().HandleFunc("/wallet-tickets", WalletTickets)

	dcrdClient.WaitForShutdown()
	dcrwalletClient1.WaitForShutdown()
	dcrwalletClient2.WaitForShutdown()
	dcrwalletClient3.WaitForShutdown()
	dcrwalletClient4.WaitForShutdown()
}

func WalletConnect(host string, user string, pass string, cert string, disableTLS bool) *rpcclient.Client {
	var certs []byte
	var err error
	certs, err = ioutil.ReadFile(cert)
	if err != nil {
		log.Printf("Failed to read wallet cert at %s: %s\n", cert, err.Error())
	}
	connConfig := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
		Certificates: certs,
		DisableTLS:   disableTLS,
	}
	walletClient, err := rpcclient.New(connConfig, nil)
	if err != nil || walletClient == nil {
		// do nothing
	}
	return walletClient
}

func (c *appContext) txGetter( w http.ResponseWriter, r *http.Request ) {
	account := GetAccountCtx(r)
	if account == "" {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	var wallets = []*rpcclient.Client
	var all bool = false
	switch account {
		case "all":
			all = true
			fallthrough
		case "cool-default":
			wallets = append(wallets, c.coolDefault)
			if !all { break }
			fallthrough
		case "cool-june-coins"
			wallets = append(wallets, c.coolJuneCoins)
			if !all { break }
			fallthrough
		case "cool-staking-rewards"
			wallets = append(wallets, c.coolStakingRewards)
			if !all { break }
			fallthrough
		case "hot-wallet"
			wallets = append(wallets, c.hotWallet)
	}
	var all_txs []dcrjson.ListTransactionsResult
	for wallet := range wallets {
		txns, err := wallet.ListTransactionsCount("*", 999999)
		if err != nil {
			log.Printf("Not able to get transactions!")
		}
		all_txs = append(all_txs, txns)
	}
	encountered := map[string]dcrjson.ListTransactionResult{}
	for i, v := range all_txs {
		encountered[all_txs[i].TxID] = v
	}
	all_txs_dedupe := make([]dcrjson.ListTransactionResult, 0, len(encountered))
	for _, v := range encountered {
		all_txs_dedupe = append(all_txs_dedupe, v)
	}
	if all_txs_dedupe == nil {
		http.Error(w, http.StatusText(422), 422)
	}
	txns := make([]dcrjson.TxRawResult, len(all_txs_dedupe))
	for _, tx := range all_txs_dedupe {
		txhash, err := chainhash.NewHashFromStr(tx.TxID)
		if err != nil {
			log.Printf("Invalid transaction hash %s", tx.TxID)
			return nil, err
		}
		rawTx, err := c.nodeClient.GetRawTransactionVerbose(txhash)
		if err != nil {
			log.Printf("GetRawTransactionVerbose failed for: %v", txhash)
			return nil, err
		}
		txns = append(txns, rawTx)
		// We now have details for all transactions from all accounts (queried)
		// TODO: Time to write it out to json
		//   TODO: Also create my own struct for the type of object I want to return
		//       so that only the fields I want are imported into Excel, so that there
		//	 is less for Excel to do, as far as horesepower goes
	}


} 

func AccountCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := chi.URLParam(r, "account")
		ctx := context.WithValue(r.Context(), ctxAccount, query)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAccountCtx(r *http.Request) string {
	account, ok := r.Context().Value(ctxAccount).(string)
	if !ok {
		log.Printf("Could not get account string from URL entered")
		return ""
	}
	return account
}


























