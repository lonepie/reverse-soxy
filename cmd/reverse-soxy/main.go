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

	// ensure shared secret is provided
	if *secretFlag == "" {
		logger.Fatal("Shared secret required: use -secret flag or config")
	}

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
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			logger.Fatalf("Failed to parse config: %v", err)
		}
		if cfg.SocksListenAddr != "" {
			*socksAddr = cfg.SocksListenAddr
		}
		if cfg.TunnelListenPort != 0 {
			*tunnelPort = cfg.TunnelListenPort
		}
		if cfg.TunnelAddr != "" {
			*tunnelAddr = cfg.TunnelAddr
		}
		logger.Debug("Loaded config from %s: socks_listen_addr=%s, tunnel_listen_port=%d, tunnel_addr=%s", *cfgPath, cfg.SocksListenAddr, cfg.TunnelListenPort, cfg.TunnelAddr)
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
		proxy.RunRelayAgent(*relayAddr, *secretFlag)
	} else if *tunnelAddr != "" {
		// direct agent
		proxy.RunAgent(*tunnelAddr, *secretFlag)
	} else {
		proxy.RunProxy(*socksAddr, *tunnelPort, *secretFlag)
	}
}
