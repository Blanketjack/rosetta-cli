// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/constructor/dsl"
	"github.com/coinbase/rosetta-sdk-go/constructor/job"
	"github.com/coinbase/rosetta-sdk-go/storage"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/coinbase/rosetta-sdk-go/utils"
	"github.com/fatih/color"
)

// CheckDataEndCondition is a type of "successful" end
// for the "check:data" method.
type CheckDataEndCondition string

const (
	// IndexEndCondition is used to indicate that the index end condition
	// has been met.
	IndexEndCondition CheckDataEndCondition = "Index End Condition"

	// DurationEndCondition is used to indicate that the duration
	// end condition has been met.
	DurationEndCondition CheckDataEndCondition = "Duration End Condition"

	// TipEndCondition is used to indicate that the tip end condition
	// has been met.
	TipEndCondition CheckDataEndCondition = "Tip End Condition"

	// ReconciliationCoverageEndCondition is used to indicate that the reconciliation
	// coverage end condition has been met.
	ReconciliationCoverageEndCondition CheckDataEndCondition = "Reconciliation Coverage End Condition"
)

// Default Configuration Values
const (
	DefaultURL                               = "http://localhost:8080"
	DefaultTimeout                           = 10
	DefaultMaxRetries                        = 5
	DefaultMaxOnlineConnections              = 120 // most OS have a default limit of 128
	DefaultMaxOfflineConnections             = 4   // we shouldn't need many connections for construction
	DefaultMaxSyncConcurrency                = 64
	DefaultActiveReconciliationConcurrency   = 16
	DefaultInactiveReconciliationConcurrency = 4
	DefaultInactiveReconciliationFrequency   = 250
	DefaultConfirmationDepth                 = 10
	DefaultStaleDepth                        = 30
	DefaultBroadcastLimit                    = 3
	DefaultTipDelay                          = 300
	DefaultBlockBroadcastLimit               = 5
	DefaultStatusPort                        = 9090

	// ETH Defaults
	EthereumIDBlockchain = "Ethereum"
	EthereumIDNetwork    = "Ropsten"
)

// Default Configuration Values
var (
	EthereumNetwork = &types.NetworkIdentifier{
		Blockchain: EthereumIDBlockchain,
		Network:    EthereumIDNetwork,
	}
)

// ConstructionConfiguration contains all configurations
// to run check:construction.
type ConstructionConfiguration struct {
	// OfflineURL is the URL of a Rosetta API implementation in "offline mode".
	OfflineURL string `json:"offline_url"`

	// MaxOffineConnections is the maximum number of open connections that the offline
	// fetcher will open.
	MaxOfflineConnections int `json:"max_offline_connections"`

	// StaleDepth is the number of blocks to wait before attempting
	// to rebroadcast after not finding a transaction on-chain.
	StaleDepth int64 `json:"stale_depth"`

	// BroadcastLimit is the number of times to attempt re-broadcast
	// before giving up on a transaction broadcast.
	BroadcastLimit int `json:"broadcast_limit"`

	// IgnoreBroadcastFailures determines if we should exit when there
	// are broadcast failures (that surpass the BroadcastLimit).
	IgnoreBroadcastFailures bool `json:"ignore_broadcast_failures"`

	// ClearBroadcasts indicates if all pending broadcasts should
	// be removed from BroadcastStorage on restart.
	ClearBroadcasts bool `json:"clear_broadcasts"`

	// BroadcastBehindTip indicates if we should broadcast transactions
	// when we are behind tip (as defined by TipDelay).
	BroadcastBehindTip bool `json:"broadcast_behind_tip"`

	// BlockBroadcastLimit is the number of transactions to attempt
	// broadcast in a single block. When there are many pending
	// broadcasts, it may make sense to limit the number of broadcasts.
	BlockBroadcastLimit int `json:"block_broadcast_limit"`

	// RebroadcastAll indicates if all pending broadcasts should be
	// rebroadcast from BroadcastStorage on restart.
	RebroadcastAll bool `json:"rebroadcast_all"`

	// PrefundedAccounts is an array of prefunded accounts
	// to use while testing.
	PrefundedAccounts []*storage.PrefundedAccount `json:"prefunded_accounts,omitempty"`

	// Workflows are executed by the rosetta-cli to test
	// certain construction flows.
	Workflows []*job.Workflow `json:"workflows"`

	// ConstructorDSLFile is the path of a Rosetta Constructor
	// DSL file (*.ros) that describes which Workflows to test.
	//
	// DSL Spec: https://github.com/coinbase/rosetta-sdk-go/tree/master/constructor/dsl
	ConstructorDSLFile string `json:"constructor_dsl_file"`

	// EndConditions is a map of workflow:count that
	// indicates how many of each workflow should be performed
	// before check:construction should stop. For example,
	// {"create_account": 5} indicates that 5 "create_account"
	// workflows should be performed before stopping.
	EndConditions map[string]int `json:"end_conditions,omitempty"`

	// StatusPort allows the caller to query a running check:construction
	// test to get stats about progress. This can be used instead
	// of parsing logs to populate some sort of status dashboard.
	StatusPort uint `json:"status_port,omitempty"`

	// ResultsOutputFile is the absolute filepath of where to save
	// the results of a check:construction run.
	ResultsOutputFile string `json:"results_output_file,omitempty"`

	// Quiet is a boolean indicating if all request and response
	// logging should be silenced.
	Quiet bool `json:"quiet,omitempty"`
}

// DefaultDataConfiguration returns the default *DataConfiguration
// for running `check:data`.
func DefaultDataConfiguration() *DataConfiguration {
	return &DataConfiguration{
		ActiveReconciliationConcurrency:   DefaultActiveReconciliationConcurrency,
		InactiveReconciliationConcurrency: DefaultInactiveReconciliationConcurrency,
		InactiveReconciliationFrequency:   DefaultInactiveReconciliationFrequency,
		StatusPort:                        DefaultStatusPort,
	}
}

// DefaultConfiguration returns a *Configuration with the
// EthereumNetwork, DefaultURL, DefaultTimeout,
// DefaultConstructionConfiguration and DefaultDataConfiguration.
func DefaultConfiguration() *Configuration {
	return &Configuration{
		Network:              EthereumNetwork,
		OnlineURL:            DefaultURL,
		MaxOnlineConnections: DefaultMaxOnlineConnections,
		HTTPTimeout:          DefaultTimeout,
		MaxRetries:           DefaultMaxRetries,
		MaxSyncConcurrency:   DefaultMaxSyncConcurrency,
		TipDelay:             DefaultTipDelay,
		Data:                 DefaultDataConfiguration(),
	}
}

// DataEndConditions contains all the conditions for the syncer to stop
// when running check:data.
type DataEndConditions struct {
	// Index configures the syncer to stop once reaching a particular block height.
	Index *int64 `json:"index,omitempty"`

	// Tip configures the syncer to stop once it reached the tip.
	// Make sure to configure `tip_delay` if you use this end
	// condition.
	Tip *bool `json:"tip,omitempty"`

	// Duration configures the syncer to stop after running
	// for Duration seconds.
	Duration *uint64 `json:"duration,omitempty"`

	// ReconciliationCoverage configures the syncer to stop
	// once it has reached tip AND some proportion of
	// all addresses have been reconciled at an index >=
	// to when tip was first reached. The range of inputs
	// for this condition are [0.0, 1.0].
	ReconciliationCoverage *float64 `json:"reconciliation_coverage,omitempty"`
}

// DataConfiguration contains all configurations to run check:data.
type DataConfiguration struct {
	// ActiveReconciliationConcurrency is the concurrency to use while fetching accounts
	// during active reconciliation.
	ActiveReconciliationConcurrency uint64 `json:"active_reconciliation_concurrency"`

	// InactiveReconciliationConcurrency is the concurrency to use while fetching accounts
	// during inactive reconciliation.
	InactiveReconciliationConcurrency uint64 `json:"inactive_reconciliation_concurrency"`

	// InactiveReconciliationFrequency is the number of blocks to wait between
	// inactive reconiliations on each account.
	InactiveReconciliationFrequency uint64 `json:"inactive_reconciliation_frequency"`

	// LogBlocks is a boolean indicating whether to log processed blocks.
	LogBlocks bool `json:"log_blocks"`

	// LogTransactions is a boolean indicating whether to log processed transactions.
	LogTransactions bool `json:"log_transactions"`

	// LogBalanceChanges is a boolean indicating whether to log all balance changes.
	LogBalanceChanges bool `json:"log_balance_changes"`

	// LogReconciliations is a boolean indicating whether to log all reconciliations.
	LogReconciliations bool `json:"log_reconciliations"`

	// IgnoreReconciliationError determines if block processing should halt on a reconciliation
	// error. It can be beneficial to collect all reconciliation errors or silence
	// reconciliation errors during development.
	IgnoreReconciliationError bool `json:"ignore_reconciliation_error"`

	// ExemptAccounts is a path to a file listing all accounts to exempt from balance
	// tracking and reconciliation. Look at the examples directory for an example of
	// how to structure this file.
	ExemptAccounts string `json:"exempt_accounts"`

	// BootstrapBalances is a path to a file used to bootstrap balances
	// before starting syncing. If this value is populated after beginning syncing,
	// it will be ignored.
	BootstrapBalances string `json:"bootstrap_balances"`

	// HistoricalBalanceEnabled is a boolean that dictates how balance lookup is performed.
	// When set to true, balances are looked up at the block where a balance
	// change occurred instead of at the current block. Blockchains that do not support
	// historical balance lookup should set this to false.
	HistoricalBalanceEnabled *bool `json:"historical_balance_enabled,omitempty"`

	// InterestingAccounts is a path to a file listing all accounts to check on each block. Look
	// at the examples directory for an example of how to structure this file.
	InterestingAccounts string `json:"interesting_accounts"`

	// ReconciliationDisabled is a boolean that indicates reconciliation should not
	// be attempted. When first testing an implementation, it can be useful to disable
	// some of the more advanced checks to confirm syncing is working as expected.
	ReconciliationDisabled bool `json:"reconciliation_disabled"`

	// InactiveDiscrepencySearchDisabled is a boolean indicating if a search
	// should be performed to find any inactive reconciliation discrepencies.
	// Note, a search will never be performed if historical balance lookup
	// is disabled.
	InactiveDiscrepencySearchDisabled bool `json:"inactive_discrepency_search_disabled"`

	// BalanceTrackingDisabled is a boolean that indicates balances calculation
	// should not be attempted. When first testing an implemenation, it can be
	// useful to just try to fetch all blocks before checking for balance
	// consistency.
	BalanceTrackingDisabled bool `json:"balance_tracking_disabled"`

	// CoinTrackingDisabled is a boolean that indicates coin (or UTXO) tracking
	// should not be attempted. When first testing an implemenation, it can be
	// useful to just try to fetch all blocks before checking for coin
	// consistency.
	CoinTrackingDisabled bool `json:"coin_tracking_disabled"`

	// StartIndex is the block height to start syncing from. If no StartIndex
	// is provided, syncing will start from the last saved block.
	// If no blocks have ever been synced, syncing will start from genesis.
	StartIndex *int64 `json:"start_index,omitempty"`

	// EndCondition contains the conditions for the syncer to stop
	EndConditions *DataEndConditions `json:"end_conditions,omitempty"`

	// StatusPort allows the caller to query a running check:data
	// test to get stats about progress. This can be used instead
	// of parsing logs to populate some sort of status dashboard.
	StatusPort uint `json:"status_port,omitempty"`

	// ResultsOutputFile is the absolute filepath of where to save
	// the results of a check:data run.
	ResultsOutputFile string `json:"results_output_file"`

	// PruningDisabled is a bolean that indicates storage pruning should
	// not be attempted. This should really only ever be set to true if you
	// wish to use `start_index` at a later point to restart from some
	// previously synced block.
	PruningDisabled bool `json:"pruning_disabled"`
}

// Configuration contains all configuration settings for running
// check:data or check:construction.
type Configuration struct {
	// Network is the *types.NetworkIdentifier where transactions should
	// be constructed and where blocks should be synced to monitor
	// for broadcast success.
	Network *types.NetworkIdentifier `json:"network"`

	// OnlineURL is the URL of a Rosetta API implementation in "online mode".
	OnlineURL string `json:"online_url"`

	// DataDirectory is a folder used to store logs and any data used to perform validation.
	DataDirectory string `json:"data_directory"`

	// HTTPTimeout is the timeout for a HTTP request in seconds.
	HTTPTimeout uint64 `json:"http_timeout"`

	// MaxRetries is the number of times we will retry an HTTP request. If retry_elapsed_time
	// is also populated, we may stop attempting retries early.
	MaxRetries uint64 `json:"max_retries"`

	// RetryElapsedTime is the total time to spend retrying a HTTP request in seconds.
	RetryElapsedTime uint64 `json:"retry_elapsed_time"`

	// MaxOnlineConnections is the maximum number of open connections that the online
	// fetcher will open.
	MaxOnlineConnections int `json:"max_online_connections"`

	// MaxSyncConcurrency is the maximum sync concurrency to use while syncing blocks.
	// Sync concurrency is managed automatically by the `syncer` package.
	MaxSyncConcurrency int64 `json:"max_sync_concurrency"`

	// TipDelay dictates how many seconds behind the current time is considered
	// tip. If we are > TipDelay seconds from the last processed block,
	// we are considered to be behind tip.
	TipDelay int64 `json:"tip_delay"`

	// LogConfiguration determines if the configuration settings
	// should be printed to the console when a file is loaded.
	LogConfiguration bool `json:"log_configuration"`

	Construction *ConstructionConfiguration `json:"construction"`
	Data         *DataConfiguration         `json:"data"`
}

func populateConstructionMissingFields(
	constructionConfig *ConstructionConfiguration,
) *ConstructionConfiguration {
	if constructionConfig == nil {
		return nil
	}

	if len(constructionConfig.OfflineURL) == 0 {
		constructionConfig.OfflineURL = DefaultURL
	}

	if constructionConfig.MaxOfflineConnections == 0 {
		constructionConfig.MaxOfflineConnections = DefaultMaxOfflineConnections
	}

	if constructionConfig.StaleDepth == 0 {
		constructionConfig.StaleDepth = DefaultStaleDepth
	}

	if constructionConfig.BroadcastLimit == 0 {
		constructionConfig.BroadcastLimit = DefaultBroadcastLimit
	}

	if constructionConfig.BlockBroadcastLimit == 0 {
		constructionConfig.BlockBroadcastLimit = DefaultBlockBroadcastLimit
	}

	if constructionConfig.StatusPort == 0 {
		constructionConfig.StatusPort = DefaultStatusPort
	}

	return constructionConfig
}

func populateDataMissingFields(dataConfig *DataConfiguration) *DataConfiguration {
	if dataConfig == nil {
		return DefaultDataConfiguration()
	}

	if dataConfig.ActiveReconciliationConcurrency == 0 {
		dataConfig.ActiveReconciliationConcurrency = DefaultActiveReconciliationConcurrency
	}

	if dataConfig.InactiveReconciliationConcurrency == 0 {
		dataConfig.InactiveReconciliationConcurrency = DefaultInactiveReconciliationConcurrency
	}

	if dataConfig.InactiveReconciliationFrequency == 0 {
		dataConfig.InactiveReconciliationFrequency = DefaultInactiveReconciliationFrequency
	}

	if dataConfig.StatusPort == 0 {
		dataConfig.StatusPort = DefaultStatusPort
	}

	return dataConfig
}

func populateMissingFields(config *Configuration) *Configuration {
	if config == nil {
		return DefaultConfiguration()
	}

	if config.Network == nil {
		config.Network = EthereumNetwork
	}

	if len(config.OnlineURL) == 0 {
		config.OnlineURL = DefaultURL
	}

	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = DefaultTimeout
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultMaxRetries
	}

	if config.MaxOnlineConnections == 0 {
		config.MaxOnlineConnections = DefaultMaxOnlineConnections
	}

	if config.MaxSyncConcurrency == 0 {
		config.MaxSyncConcurrency = DefaultMaxSyncConcurrency
	}

	if config.TipDelay == 0 {
		config.TipDelay = DefaultTipDelay
	}

	config.Construction = populateConstructionMissingFields(config.Construction)
	config.Data = populateDataMissingFields(config.Data)

	return config
}

func assertConstructionConfiguration(ctx context.Context, config *ConstructionConfiguration) error {
	if config == nil {
		return nil
	}

	if len(config.Workflows) > 0 && len(config.ConstructorDSLFile) > 0 {
		return errors.New("cannot populate both workflows and DSL file path")
	}

	if len(config.Workflows) == 0 && len(config.ConstructorDSLFile) == 0 {
		return errors.New("both workflows and DSL file path are empty")
	}

	// Compile ConstructorDSLFile and save to Workflows
	if len(config.ConstructorDSLFile) > 0 {
		compiledWorkflows, err := dsl.Parse(ctx, config.ConstructorDSLFile)
		if err != nil {
			return fmt.Errorf("%s: compilation failed", types.PrintStruct(err))
		}

		config.Workflows = compiledWorkflows
	}

	// Parse provided Workflows
	for _, workflow := range config.Workflows {
		if workflow.Name == string(job.CreateAccount) || workflow.Name == string(job.RequestFunds) {
			if workflow.Concurrency != job.ReservedWorkflowConcurrency {
				return fmt.Errorf(
					"reserved workflow %s must have concurrency %d",
					workflow.Name,
					job.ReservedWorkflowConcurrency,
				)
			}
		}
	}

	for _, account := range config.PrefundedAccounts {
		// Checks that privkey is hex encoded
		_, err := hex.DecodeString(account.PrivateKeyHex)
		if err != nil {
			return fmt.Errorf(
				"%w: private key %s is not hex encoded for prefunded account",
				err,
				account.PrivateKeyHex,
			)
		}

		// Checks if valid CurveType
		if err := asserter.CurveType(account.CurveType); err != nil {
			return fmt.Errorf("%w: invalid CurveType for prefunded account", err)
		}

		// Checks if valid AccountIdentifier
		if err := asserter.AccountIdentifier(account.AccountIdentifier); err != nil {
			return fmt.Errorf("Account.Address is missing for prefunded account")
		}

		// Check if valid Currency
		err = asserter.Currency(account.Currency)
		if err != nil {
			return fmt.Errorf("%w: invalid currency for prefunded account", err)
		}
	}

	return nil
}

func assertDataConfiguration(config *DataConfiguration) error {
	if config.StartIndex != nil && *config.StartIndex < 0 {
		return fmt.Errorf("start index %d cannot be negative", *config.StartIndex)
	}

	if config.EndConditions == nil {
		return nil
	}

	if config.EndConditions.Index != nil {
		if *config.EndConditions.Index < 0 {
			return fmt.Errorf("end index %d cannot be negative", *config.EndConditions.Index)
		}
	}

	if config.EndConditions.ReconciliationCoverage != nil {
		coverage := *config.EndConditions.ReconciliationCoverage
		if coverage < 0 || coverage > 1 {
			return fmt.Errorf("reconciliation coverage %f must be [0.0,1.0]", coverage)
		}

		if config.BalanceTrackingDisabled {
			return errors.New(
				"balance tracking must be enabled for reconciliation coverage end condition",
			)
		}

		if config.IgnoreReconciliationError {
			return errors.New(
				"reconciliation errors cannot be ignored for reconciliation coverage end condition",
			)
		}

		if config.ReconciliationDisabled {
			return errors.New(
				"reconciliation cannot be disabled for reconciliation coverage end condition",
			)
		}
	}

	return nil
}

func assertConfiguration(ctx context.Context, config *Configuration) error {
	if err := asserter.NetworkIdentifier(config.Network); err != nil {
		return fmt.Errorf("%w: invalid network identifier", err)
	}

	if err := assertDataConfiguration(config.Data); err != nil {
		return fmt.Errorf("%w: invalid data configuration", err)
	}

	if err := assertConstructionConfiguration(ctx, config.Construction); err != nil {
		return fmt.Errorf("%w: invalid construction configuration", err)
	}

	return nil
}

// LoadConfiguration returns a parsed and asserted Configuration for running
// tests.
func LoadConfiguration(ctx context.Context, filePath string) (*Configuration, error) {
	var configRaw Configuration
	if err := utils.LoadAndParse(filePath, &configRaw); err != nil {
		return nil, fmt.Errorf("%w: unable to open configuration file", err)
	}

	config := populateMissingFields(&configRaw)

	if err := assertConfiguration(ctx, config); err != nil {
		return nil, fmt.Errorf("%w: invalid configuration", err)
	}

	color.Cyan(
		"loaded configuration file: %s\n",
		filePath,
	)

	if config.LogConfiguration {
		log.Println(types.PrettyPrintStruct(config))
	}

	return config, nil
}
