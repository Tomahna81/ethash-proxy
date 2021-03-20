package proxy

import (
	"github.com/Tomahna81/ethash-mining-stratum-proxy/api"
	"github.com/Tomahna81/ethash-mining-stratum-proxy/exchange"
	"github.com/Tomahna81/ethash-mining-stratum-proxy/payouts"
	"github.com/Tomahna81/ethash-mining-stratum-proxy/policy"
	"github.com/Tomahna81/ethash-mining-stratum-proxy/storage"
)

type Config struct {
	Name                  string        `json:"name"`
	Proxy                 Proxy         `json:"proxy"`
	Api                   api.ApiConfig `json:"api"`
	Upstream              []Upstream    `json:"upstream"`
	UpstreamCheckInterval string        `json:"upstreamCheckInterval"`

	Threads int `json:"threads"`

	Coin     string         `json:"coin"`
	Pplns    int64          `json:"pplns"`
	CoinName string         `json:"coin-name"`
	Redis    storage.Config `json:"redis"`

	BlockUnlocker payouts.UnlockerConfig `json:"unlocker"`
	Payouts       payouts.PayoutsConfig  `json:"payouts"`

	Exchange exchange.ExchangeConfig `json:"exchange"`

	NewrelicName    string `json:"newrelicName"`
	NewrelicKey     string `json:"newrelicKey"`
	NewrelicVerbose bool   `json:"newrelicVerbose"`
	NewrelicEnabled bool   `json:"newrelicEnabled"`
}

type Proxy struct {
	Enabled              bool   `json:"enabled"`
	Username             string `json:"user"`
	Password             string `json:"password"`
	Listen               string `json:"listen"`
	LimitHeadersSize     int    `json:"limitHeadersSize"`
	LimitBodySize        int64  `json:"limitBodySize"`
	BehindReverseProxy   bool   `json:"behindReverseProxy"`
	BlockRefreshInterval string `json:"blockRefreshInterval"`
	Difficulty           int64  `json:"difficulty"`
	StateUpdateInterval  string `json:"stateUpdateInterval"`
	HashrateExpiration   string `json:"hashrateExpiration"`

	Policy policy.Config `json:"policy"`

	MaxFails    int64 `json:"maxFails"`
	HealthCheck bool  `json:"healthCheck"`

	Stratum    Stratum    `json:"stratum"`
	StratumSSL StratumSSL `json:"stratum_ssl"`

	StratumNiceHash StratumNiceHash `json:"stratum_nice_hash"`
}

type Stratum struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
	Timeout string `json:"timeout"`
	MaxConn int    `json:"maxConn"`
}

type StratumSSL struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Timeout  string `json:"timeout"`
	MaxConn  int    `json:"maxConn"`
	CertFile string `json:"certfile"`
	CertKey  string `json:"certkey"`
}

type StratumNiceHash struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
	Timeout string `json:"timeout"`
	MaxConn int    `json:"maxConn"`
}

type Upstream struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	Port    string  `json:"port"`
	Timeout string `json:"timeout"`
	Username string `json:"username"`
}
