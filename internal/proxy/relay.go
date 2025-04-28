package proxy

import (
    "fmt"
    "io"
    "net"
    "sync"
    "strings"
    "time"

    "github.com/lonepie/reverse-soxy/internal/logger"
)

var (
    regMu    sync.Mutex
    registry []net.Conn
)

// RunRelay starts a relay server on the given port, accepting registrations and agent connections.
func RunRelay(listenPort int, secret string) {
    addr := fmt.Sprintf(":%d", listenPort)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        logger.Fatalf("Relay listen error: %v", err)
    }
    logger.Info("Relay listening on %s", addr)
    for {
        conn, err := listener.Accept()
        if err != nil {
            logger.Error("Relay accept error: %v", err)
            continue
        }
        go handleRelayConn(conn, secret)
    }
}

func handleRelayConn(conn net.Conn, secret string) {
    defer conn.Close()
    hdr := make([]byte, 8)
    if _, err := io.ReadFull(conn, hdr); err != nil {
        logger.Error("Relay header read error: %v", err)
        return
    }
    header := strings.TrimSpace(string(hdr))
    switch header {
    case "REGISTER":
        registerProxy(conn)
    case "AGENT":
        handleAgent(conn)
    default:
        logger.Error("Unknown relay header: %s", header)
    }
}

func registerProxy(conn net.Conn) {
    regMu.Lock()
    registry = append(registry, conn)
    regMu.Unlock()
    logger.Info("Proxy registered to relay")
    select {} // hold open
}

func handleAgent(conn net.Conn) {
    regMu.Lock()
    if len(registry) == 0 {
        regMu.Unlock()
        logger.Error("No registered proxies available")
        return
    }
    proxyConn := registry[0]
    registry = registry[1:]
    regMu.Unlock()
    go io.Copy(proxyConn, conn)
    io.Copy(conn, proxyConn)
}

// RunRegister connects to a relay server and registers as a proxy.
func RunRegister(relayAddr, secret string) {
    conn, err := net.Dial("tcp", relayAddr)
    if err != nil {
        logger.Fatalf("Register dial error: %v", err)
    }
    defer conn.Close()
    conn.Write([]byte("REGISTER"))
    logger.Info("Registered with relay %s", relayAddr)
    select {} // hold open
}

// RunRelayAgent dials into the relay server, sends AGENT header, then initiates secure tunnel
func RunRelayAgent(relayAddr, secret string) {
    for {
        rawConn, err := net.Dial("tcp", relayAddr)
        if err != nil {
            logger.Error("RelayAgent dial failed: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        // send fixed 8-byte header
        hdr := []byte("AGENT   ")
        if _, err := rawConn.Write(hdr); err != nil {
            rawConn.Close()
            logger.Error("RelayAgent header send error: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }
        // secure handshake
        secureConn, err := NewSecureClientConn(rawConn, secret)
        if err != nil {
            logger.Error("RelayAgent handshake failed: %v", err)
            rawConn.Close()
            time.Sleep(5 * time.Second)
            continue
        }
        logger.Info("RelayAgent connected via relay %s", relayAddr)
        // handle tunnel until error
        handleTunnelReadsServer(secureConn)
        secureConn.Close()
        time.Sleep(5 * time.Second)
    }
}
