package main

import (
	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/Rudd-O/curvetls"
	"github.com/anxiousmodernman/hpt/proto/server"
)

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
	creds := curvetls.NewGRPCClientCredentials(tpub, pub, priv)
	conn, err := grpc.Dial(targetAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrapf(err, "dialing %s", targetAddr)
	}

	return server.NewHPTClient(conn), nil
}
