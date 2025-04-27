package proxy

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type session struct {
	targetConn net.Conn
	incoming   chan []byte
	ready      chan struct{}
}

var (
	sessionsMu  sync.Mutex
	sessionsMap = make(map[uint32]*session)
)

// RunServer starts the proxy server connecting over the tunnel to the client.
func RunServer(tunnelAddr string) {
	tunnel, err := net.Dial("tcp", tunnelAddr)
	if err != nil {
		log.Fatal("[SERVER] Tunnel connection failed:", err)
	}
	if tcpConn, ok := tunnel.(*net.TCPConn); ok {
		err := tcpConn.SetNoDelay(true)
		if err != nil {
			log.Printf("[SERVER] Failed to set TCP_NODELAY: %v", err)
		} else {
			log.Println("[SERVER] Set TCP_NODELAY (NoDelay) on tunnel connection")
		}
	}
	log.Println("[SERVER] Tunnel connected to laptop")
	// Only handleTunnelReadsServer reads from the tunnel connection.
	handleTunnelReadsServer(tunnel)
}

func handleTunnelReadsServer(tunnel net.Conn) {
	log.Println("[SERVER] Starting to read from tunnel")
	for {
		log.Println("[SERVER] Waiting for tunnel header...")
		header := make([]byte, 6)
		_, err := io.ReadFull(tunnel, header)
		if err != nil {
			log.Println("[SERVER] tunnel read error:", err)
			return
		}
		sessID := binary.BigEndian.Uint32(header[:4])
		log.Printf("[SERVER] Read header - session ID: %08x", sessID)
		length := binary.BigEndian.Uint16(header[4:6])
		log.Printf("[SERVER] Expected payload length: %d", length)

		sessionsMu.Lock()
		sess, ok := sessionsMap[sessID]
		sessionsMu.Unlock()
		if !ok {
			// First message for this session: target string
			targetBuf := make([]byte, length)
			_, err = io.ReadFull(tunnel, targetBuf)
			if err != nil {
				log.Println("[SERVER] Failed to read target address:", err)
				continue
			}
			target := string(targetBuf)
			log.Printf("[SERVER] session %08x connecting to %s", sessID, target)

			tgtConn, err := net.Dial("tcp", target)
			if err != nil {
				log.Printf("[SERVER] session %08x dial failed: %v", sessID, err)
				continue
			}
			sess = &session{
				targetConn: tgtConn,
				incoming:   make(chan []byte, 10),
				ready:      make(chan struct{}),
			}
			sessionsMu.Lock()
			sessionsMap[sessID] = sess
			sessionsMu.Unlock()
			close(sess.ready)
			go handleSession(sessID, sess, tunnel)
			continue // go to next header, do not expect payload for this header
		}

		// Existing session: treat as payload
		select {
		case <-sess.ready:
			log.Printf("[SERVER] session %08x is ready to receive payload", sessID)
		case <-time.After(2 * time.Second):
			log.Printf("[SERVER] session %08x timed out waiting for ready", sessID)
			continue
		}

		log.Printf("[SERVER] about to read %d bytes of payload for session %08x", length, sessID)
		buf := make([]byte, length)
		bytesRead := 0
		for bytesRead < int(length) {
			n, err := tunnel.Read(buf[bytesRead:])
			log.Printf("[SERVER][DEBUG] Read returned %d bytes, err=%v", n, err)
			log.Printf("[SERVER][DEBUG] Bytes read so far (%d/%d): %x", bytesRead+n, length, buf[:bytesRead+n])
			if n > 0 {
				bytesRead += n
			}
			if err != nil {
				log.Printf("[SERVER] session %08x payload read error after %d bytes: %v", sessID, bytesRead, err)
				return
			}
		}
		log.Printf("[SERVER] finished reading payload for session %08x", sessID)
		log.Printf("[SERVER] session %08x received %d bytes from tunnel Payload (hex): %x", sessID, length, buf)
		sess.incoming <- buf
	}
}

var serverTunnelWriteMu sync.Mutex

func handleSession(sessID uint32, sess *session, tunnel net.Conn) {
	defer sess.targetConn.Close()
	defer func() {
		sessionsMu.Lock()
		delete(sessionsMap, sessID)
		sessionsMu.Unlock()
		close(sess.incoming)
	}()

	go func() {
		buf := make([]byte, 4096)
		header := make([]byte, 6)
		binary.BigEndian.PutUint32(header[:4], sessID)
		for {
			n, err := sess.targetConn.Read(buf)
			if err != nil {
				log.Printf("[SERVER] session %08x closed by target\n", sessID)
				return
			}
			binary.BigEndian.PutUint16(header[4:], uint16(n))
			serverTunnelWriteMu.Lock()
			tunnel.Write(header)
			tunnel.Write(buf[:n])
			serverTunnelWriteMu.Unlock()
			log.Printf("[SERVER] session %08x sent %d bytes to tunnel\nPayload: %s\n", sessID, n, string(buf[:n]))
		}
	}()

	for data := range sess.incoming {
		log.Printf("[SERVER] session %08x writing %d bytes to target", sessID, len(data))
		n, err := sess.targetConn.Write(data)
		if err != nil {
			log.Printf("[SERVER] session %08x write to target failed: %v", sessID, err)
			break
		}
		log.Printf("[SERVER] session %08x forwarded %d bytes to target", sessID, n)
	}
}
