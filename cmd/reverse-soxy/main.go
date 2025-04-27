package main

import (
	"flag"
	"math/rand"
	"os"
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
	socksAddr := flag.String("socks-listen-addr", "127.0.0.1:1080", "SOCKS5 listen address")
	tunnelPort := flag.Int("tunnel-listen-port", 9000, "Tunnel listen port when in SOCKS-frontend (default listener)")
	dialMode := flag.Bool("dial", false, "Dial tunnel to remote (client mode)")
	dialAddr := flag.String("dial-addr", "", "Remote tunnel address (peer IP:port) when dialing")
	secretFlag := flag.String("secret", "", "shared secret for tunnel encryption/authentication")
	cfgPath := flag.String("config", "", "YAML config file path")
	flag.Parse()

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
			*dialAddr = cfg.TunnelAddr
		}
		logger.Debug("Loaded config from %s: socks_listen_addr=%s, tunnel_listen_port=%d, tunnel_addr=%s", *cfgPath, cfg.SocksListenAddr, cfg.TunnelListenPort, cfg.TunnelAddr)
	}

	// Determine component (CLIENT or SERVER)
	role := "CLIENT"
	if *dialMode {
		role = "CLIENT"
	} else {
		role = "SERVER"
	}
	logger.Init(*debugFlag, role)
	logger.Info("Debug logging enabled: %v", *debugFlag)

	// Dispatch
	logger.Debug("CLI flags: socks-listen-addr=%s, tunnel-listen-port=%d, dial=%v, dial-addr=%s, secret=%s, config=%s", *socksAddr, *tunnelPort, *dialMode, *dialAddr, *secretFlag, *cfgPath)
	if *dialMode {
		if *dialAddr == "" {
			logger.Fatal("Remote tunnel address required in dial mode")
		}
		proxy.RunTunnelDialer(*dialAddr, *secretFlag)
	} else {
		proxy.RunSOCKSFrontend(*socksAddr, *tunnelPort, *secretFlag)
	}
}
