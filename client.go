package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Rudd-O/curvetls"
	"github.com/anxiousmodernman/hpt/proto/server"
)

type HPTClient struct {
}

// NewHPTClient returns our client that can connect to remote targets over gRPC.
func NewHPTClient(keystorePath, targetName, targetAddr string) (server.HPTClient, error) {
	db, err := storm.Open(keystorePath)
	if err != nil {
		return nil, err
	}
	// ours: the client keypair
	var ckp KeyPair
	if err := db.One("Name", "client", &ckp); err != nil {
		return nil, err
	}

	// the target's keypair
	var tkp KeyPair
	if err := db.One("Name", targetName, &tkp); err != nil {
		return nil, errors.Wrap(err, "error get target keypair")
	}

	tpub, err := curvetls.PubkeyFromString(tkp.Pub)
	if err != nil {
		return nil, err
	}

	pub, err := curvetls.PubkeyFromString(ckp.Pub)
	if err != nil {
		return nil, err
	}
	priv, err := curvetls.PrivkeyFromString(ckp.Priv)
	if err != nil {
		return nil, err
	}
	fmt.Println("about to dial")
	creds := curvetls.NewGRPCClientCredentials(tpub, pub, priv)
	//grpc.WithDialer
	conn, err := grpc.Dial(targetAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrapf(err, "dialing %s", targetAddr)
	}
	st := conn.GetState()
	fmt.Println("conn state", st)

	return server.NewHPTClient(conn), nil
}

func doclient() {
	if len(os.Args) < 5 {
		log.Fatalf("usage: curvetls-client <IP:port> <client privkey> <client pubkey> <server pubkey>")
	}

	addr := os.Args[1]
	clientPrivkey, err := curvetls.PrivkeyFromString(os.Args[2])
	if err != nil {
		log.Fatalf("Client: failed to parse client private key: %s", err)
	}
	clientPubkey, err := curvetls.PubkeyFromString(os.Args[3])
	if err != nil {
		log.Fatalf("Client: failed to parse client public key: %s", err)
	}
	serverPubkey, err := curvetls.PubkeyFromString(os.Args[4])
	if err != nil {
		log.Fatalf("Client: failed to parse server public key: %s", err)
	}

	socket, err := net.Dial("tcp4", addr)
	if err != nil {
		log.Fatalf("Client: failed to connect to socket: %s", err)
	}

	nonce, err := curvetls.NewLongNonce()
	if err != nil {
		log.Fatalf("Failed to generate nonce: %s", err)
	}
	ssocket, err := curvetls.WrapClient(socket, clientPrivkey, clientPubkey, serverPubkey, nonce)
	if err != nil {
		if curvetls.IsAuthenticationError(err) {
			log.Fatalf("Client: server says unauthorized: %s", err)
		} else {
			log.Fatalf("Client: failed to wrap socket: %s", err)
		}
	}

	if err == nil {
		_, err = ssocket.Write([]byte("ghi jkl"))
		if err != nil {
			log.Fatalf("Client: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Client: wrote ghi jkl to wrapped socket")

		var packet [8]byte
		var smallPacket [8]byte

		_, err = ssocket.Read(packet[:])
		if err != nil {
			log.Fatalf("Client: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Client: the first received packet is %s", packet)

		_, err = ssocket.Write([]byte("GHI JKL STU VWX "))
		if err != nil {
			log.Fatalf("Client: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Client: wrote GHI JKL STU VWX to wrapped socket")

		n, err := ssocket.Read(smallPacket[:])
		if err != nil {
			log.Fatalf("Client: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Client: the second received first part of packet is %s", smallPacket[:n])

		n, err = ssocket.Read(smallPacket[:])
		if err != nil {
			log.Fatalf("Server: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Client: the second received second part of packet is %s", smallPacket[:n])

		_, err = ssocket.Write([]byte("SHORT"))
		if err != nil {
			log.Fatalf("Client: failed to write to wrapped socket: %s", err)
		}

		log.Printf("Client: wrote SHORT to wrapped socket")

		short, err := ssocket.ReadFrame()
		if err != nil {
			log.Fatalf("Client: failed to read from wrapped socket: %s", err)
		}

		log.Printf("Client: the frame received is %s", short)

		err = ssocket.Close()
		if err != nil {
			log.Fatalf("Client: failed to close socket: %s", err)
		}
	}
}
