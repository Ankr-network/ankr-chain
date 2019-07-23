package config

import (
	tmcoreconfig "github.com/tendermint/tendermint/config"
)

type HistoryDBConfig struct {
	Type string
	Host string
	Name string
}

type AnkrConfig struct {
	// Top level options use an anonymous struct
	tmcoreconfig.BaseConfig `mapstructure:",squash"`

	// Options for services
	RPC             *tmcoreconfig.RPCConfig             `mapstructure:"rpc"`
	P2P             *tmcoreconfig.P2PConfig             `mapstructure:"p2p"`
	Mempool         *tmcoreconfig.MempoolConfig         `mapstructure:"mempool"`
	Consensus       *tmcoreconfig.ConsensusConfig       `mapstructure:"consensus"`
	TxIndex         *tmcoreconfig.TxIndexConfig         `mapstructure:"tx_index"`
	Instrumentation *tmcoreconfig.InstrumentationConfig `mapstructure:"instrumentation"`
	HistoryDB       *HistoryDBConfig
}

func (ac *AnkrConfig) SetRoot(root string) *AnkrConfig {
	ac.BaseConfig.RootDir = root
	ac.RPC.RootDir = root
	ac.P2P.RootDir = root
	ac.Mempool.RootDir = root
	ac.Consensus.RootDir = root
	return ac
}

func (ac *AnkrConfig) TendermintCoreConfig() *tmcoreconfig.Config {
	return &tmcoreconfig.Config  {
		ac.BaseConfig,
		ac.RPC,
		ac.P2P,
		ac.Mempool,
		ac.Consensus,
		ac.TxIndex,
		ac.Instrumentation,
	}
}

func DefaultAnkrChainDir() string {
	return ".ankrchain"
}

func DefaultLogLevel() string {
	return "error"
}

func DefaultHistoryDBConfig() *HistoryDBConfig {
	return &HistoryDBConfig{
		Type:  "",
		Host:  "",
		Name:  "",
	}
}

func DefaultAnkrConfig() *AnkrConfig {
	tmcoreConfigBasic := tmcoreconfig.DefaultBaseConfig()
	tmcoreConfigBasic.ProxyApp = "AnkrChain"

	return &AnkrConfig {
		tmcoreConfigBasic,
		 tmcoreconfig.DefaultRPCConfig(),
		  tmcoreconfig.DefaultP2PConfig(),
		tmcoreconfig.DefaultMempoolConfig(),
		tmcoreconfig.DefaultConsensusConfig(),
		tmcoreconfig.DefaultTxIndexConfig(),
		tmcoreconfig.DefaultInstrumentationConfig(),
		DefaultHistoryDBConfig(),
	}
}






