package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/lonepie/reverse-soxy/internal/proxy"
	"gopkg.in/yaml.v3"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// CLI flags
	socksAddr := flag.String("socks", "127.0.0.1:1080", "SOCKS5 listen address")
	tunnelPort := flag.Int("tunnel-port", 9000, "Tunnel listen port (client)")
	serverMode := flag.Bool("server", false, "Run in server mode")
	tunnelAddr := flag.String("tunnel-addr", "", "Tunnel address (client IP:port) for server mode")
	cfgPath := flag.String("config", "", "YAML config file path")
	flag.Parse()

	// Optional YAML config override
	if *cfgPath != "" {
		data, err := os.ReadFile(*cfgPath)
		if err != nil {
			log.Fatalf("Failed to read config: %v", err)
		}
		var cfg struct {
			SocksListenAddr  string `yaml:"socks_listen_addr"`
			TunnelListenPort int    `yaml:"tunnel_listen_port"`
			TunnelAddr       string `yaml:"tunnel_addr"`
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("Failed to parse config: %v", err)
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
	}

	// Dispatch
	if *serverMode {
		if *tunnelAddr == "" {
			log.Fatal("Tunnel address required in server mode")
		}
		proxy.RunServer(*tunnelAddr)
	} else {
		proxy.RunClient(*socksAddr, *tunnelPort)
	}
}
