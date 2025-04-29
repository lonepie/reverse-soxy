package main

import (
	"context"
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lonepie/reverse-soxy/internal/logger"
	"github.com/lonepie/reverse-soxy/internal/proxy"
	"gopkg.in/yaml.v3"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Debug flag
	debugFlag := flag.Bool("debug", false, "enable debug logging")

	// CLI flags
	socksAddr := flag.String("proxy-listen-addr", "127.0.0.1:1080", "SOCKS5 listen address")
	tunnelPort := flag.Int("tunnel-listen-port", 9000, "Tunnel listen port when in proxy mode")
	tunnelAddr := flag.String("tunnel-addr", "", "Tunnel address (IP:port) to dial (agent mode)")
	secretFlag := flag.String("secret", "", "shared secret for tunnel encryption/authentication")
	cfgPath := flag.String("config", "", "YAML config file path")
	modeFlag := flag.String("mode", "", "Component mode: proxy (default), agent, relay")
	relayListenPort := flag.Int("relay-listen-port", 9000, "Port for both Proxy registrations and Agent tunnels (relay mode)")
	retryFlag := flag.Int("retry", 10, "Maximum number of retries")
	registerFlag := flag.Bool("register", false, "Proxy registers its availability to Relay server")
	relayAddr := flag.String("relay-addr", "", "Relay server address (IP:port) for registration or agent dialing")
	flag.Parse()

	// graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		logger.Info("Shutdown signal received, exiting")
		os.Exit(0)
	}()

	// Optional YAML config override
	if *cfgPath != "" {
		data, err := os.ReadFile(*cfgPath)
		if err != nil {
			logger.Fatalf("Failed to read config: %v", err)
		}
		var cfg struct {
			SocksListenAddr  string `yaml:"socks_listen_addr"`
			TunnelListenPort int    `yaml:"tunnel_listen_port"`
			TunnelAddr       string `yaml:"tunnel_addr"`
			MaxRetries       int    `yaml:"max_retries"`
			Secret           string `yaml:"secret"`
			RelayListenPort  int    `yaml:"relay_listen_port"`
			RelayAddr        string `yaml:"relay_addr"`
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			logger.Fatalf("Failed to parse config: %v", err)
		}

		// Apply config values only if CLI flags are at their defaults
		if *socksAddr == "127.0.0.1:1080" && cfg.SocksListenAddr != "" {
			*socksAddr = cfg.SocksListenAddr
		}
		if *tunnelPort == 9000 && cfg.TunnelListenPort != 0 {
			*tunnelPort = cfg.TunnelListenPort
		}
		if *tunnelAddr == "" && cfg.TunnelAddr != "" {
			*tunnelAddr = cfg.TunnelAddr
		}
		if *secretFlag == "" && cfg.Secret != "" {
			*secretFlag = cfg.Secret
		}
		if *relayListenPort == 9000 && cfg.RelayListenPort != 0 {
			*relayListenPort = cfg.RelayListenPort
		}
		if *relayAddr == "" && cfg.RelayAddr != "" {
			*relayAddr = cfg.RelayAddr
		}
		if *retryFlag == 10 && cfg.MaxRetries != 0 {
			*retryFlag = cfg.MaxRetries
		}

		logger.Debug("Loaded config from %s: socks_listen_addr=%s, tunnel_listen_port=%d, tunnel_addr=%s, secret=%s, relay_listen_port=%d, relay_addr=%s, max_retries=%d",
			*cfgPath,
			cfg.SocksListenAddr,
			cfg.TunnelListenPort,
			cfg.TunnelAddr,
			cfg.Secret,
			cfg.RelayListenPort,
			cfg.RelayAddr,
			cfg.MaxRetries)
	}

	// ensure shared secret is provided
	if *secretFlag == "" {
		logger.Fatal("Shared secret required: use -secret flag or config")
	}

	// validate tunnelAddr if provided
	if *tunnelAddr != "" {
		if _, _, err := net.SplitHostPort(*tunnelAddr); err != nil {
			logger.Fatalf("Invalid tunnel-addr: %v", err)
		}
	}

	// Determine component: AGENT if tunnelAddr given, else PROXY
	var role string
	if *modeFlag != "" {
		role = *modeFlag
	} else if *tunnelAddr != "" {
		role = "AGENT"
	} else if *registerFlag {
		role = "REGISTER"
	} else {
		role = "PROXY"
	}
	logger.Init(*debugFlag, role)
	logger.Info("Debug logging enabled: %v", *debugFlag)

	// Dispatch
	logger.Debug("CLI flags: proxy-listen-addr=%s, tunnel-listen-port=%d, tunnel-addr=%s, secret=%s, config=%s, mode=%s, relay-listen-port=%d, register=%v, relay-addr=%s", *socksAddr, *tunnelPort, *tunnelAddr, *secretFlag, *cfgPath, *modeFlag, *relayListenPort, *registerFlag, *relayAddr)
	if *modeFlag == "relay" {
		proxy.RunRelay(*relayListenPort, *secretFlag)
	} else if *registerFlag {
		// register with relay and start proxy via relay
		proxy.RunProxyRelay(*relayAddr, *socksAddr, *secretFlag)
	} else if *relayAddr != "" {
		// agent via relay
		proxy.RunAgentRelay(*relayAddr, *secretFlag, *retryFlag)
	} else if *tunnelAddr != "" {
		// direct agent
		proxy.RunAgent(*tunnelAddr, *secretFlag, *retryFlag)
	} else {
		proxy.RunProxy(*socksAddr, *tunnelPort, *secretFlag)
	}
}
