package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/Rudd-O/curvetls"
)

const hptsock = "hpt.sock"

func doserver() {

	var addr string

	var l net.Listener

	// get our listener from systemd, or bind to a configured address
	if os.Getenv("LISTEN_PID") == strconv.Itoa(os.Getpid()) {
		// systemd run
		var err error
		f := os.NewFile(3, hptsock)
		l, err = net.FileListener(f)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// manual run
		var err error
		l, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(os.Args) < 4 || len(os.Args) > 5 {
		log.Fatalf("usage: curvetls-server <IP:port> <server privkey> <server pubkey> [client pubkey]")
	}

	serverPrivkey, err := curvetls.PrivkeyFromString(os.Args[2])
	if err != nil {
		log.Fatalf("Server: failed to parse server private key: %s", err)
	}
	serverPubkey, err := curvetls.PubkeyFromString(os.Args[3])
	if err != nil {
		log.Fatalf("Server: failed to parse server public key: %s", err)
	}
	var noPubkey curvetls.Pubkey
	var clientPubkey curvetls.Pubkey
	if len(os.Args) == 5 {
		clientPubkey, err = curvetls.PubkeyFromString(os.Args[4])
		if err != nil {
			log.Fatalf("Server: failed to parse client public key: %s", err)
		}
	} else {
		clientPubkey = noPubkey
	}

	listener := l
	if err != nil {
		log.Fatalf("Server: could not run server: %s", err)
	}

	socket, err := listener.Accept()
	if err != nil {
		log.Fatalf("Server: failed to accept socket: %s", err)
	}

	long_nonce, err := curvetls.NewLongNonce()
	if err != nil {
		log.Fatalf("Server: failed to generate nonce: %s", err)
	}
	authorizer, clientpubkey, err := curvetls.WrapServer(socket, serverPrivkey, serverPubkey, long_nonce)
	if err != nil {
		log.Fatalf("Server: failed to wrap socket: %s", err)
	}
	log.Printf("Server: client's public key is %s", clientpubkey)

	var ssocket *curvetls.EncryptedConn

	var allowed bool
	if clientPubkey == noPubkey {
		ssocket, err = authorizer.Allow()
		allowed = true
	} else if clientPubkey == clientpubkey {
		ssocket, err = authorizer.Allow()
		allowed = true
	} else {
		err = authorizer.Deny()
		allowed = false
	}

	if err != nil {
		log.Fatalf("Server: failed to process authorization: %s", err)
	}

	if allowed {
		var packet [8]byte
		var smallPacket [8]byte

		_, err = ssocket.Read(packet[:])
		if err != nil {
			log.Fatalf("Server: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Server: the first received packet is %s", packet)

		_, err = ssocket.Write([]byte("abc def"))
		if err != nil {
			log.Fatalf("Server: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Server: wrote abc def to wrapped socket")

		n, err := ssocket.Read(smallPacket[:])
		if err != nil {
			log.Fatalf("Server: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Server: the second received first part of packet is %s", smallPacket[:n])

		n, err = ssocket.Read(smallPacket[:])
		if err != nil {
			log.Fatalf("Server: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Server: the second received second part of packet is %s", smallPacket[:n])

		_, err = ssocket.Write([]byte("ABC DEF MNO PQR"))
		if err != nil {
			log.Fatalf("Server: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Server: wrote ABC DEF MNO PQR to wrapped socket")

		short, err := ssocket.ReadFrame()
		if err != nil {
			log.Fatalf("Server: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Server: the frame received is %s", short)

		_, err = ssocket.Write([]byte("SHORT"))
		if err != nil {
			log.Fatalf("Server: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Server: wrote SHORT to wrapped socket")

		err = ssocket.Close()
		if err != nil {
			log.Fatalf("Server: failed to close socket: %s", err)
		}
	}
}
