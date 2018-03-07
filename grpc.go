package main

import (
	"context"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/Rudd-O/curvetls"
	"github.com/anxiousmodernman/hpt/proto/server"
	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var _ server.HPTServer = (*HPTServer)(nil)

// HPTServer ...
type HPTServer struct {
	// gRPC Server
	*grpc.Server
}

func NewHPTServer(path string) (*HPTServer, error) {

	// We expect the db to exist. During initial setup, an authorized client
	// will create this DB for us and populate it with authorized client keys,
	// as well as our own key. We must have all authorized clients in the db,
	// because we cannot tolerate even accepting connections from unauthorized
	// clients.
	db, err := storm.Open(path)
	if err != nil {
		return nil, err
	}

	// The curvetls interface from our fork.
	var keystore curvetls.KeyStore
	keystore = &StormKeystore{db}
	// curvetls transport security
	serverPub, serverPriv, err := RetrieveServerKeys(db)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve server's keypair %v", err)
	}
	creds := curvetls.NewGRPCServerCredentials(serverPub, serverPriv, keystore)

	svr := grpc.NewServer(grpc.Creds(creds))
	var h HPTServer
	server.RegisterHPTServer(svr, &h)
	return &h, nil
}

// Apply ...
func (h *HPTServer) Apply(conf *server.Config, stream server.HPT_ApplyServer) error {

	var c Config
	err := toml.Unmarshal(conf.Data, &c)
	if err != nil {
		return err
	}

	ep, err := NewExecutionPlan(c)
	if err != nil {
		return err
	}
	for {
		fn := ep.Next()
		if fn == nil {
			break
		}
		state := fn()
		outcome := whatHappened(state.Outcome)

		result := server.ApplyResult{
			Msg: &server.ApplyResult_Metadata{
				Metadata: &server.ApplyResultMetadata{
					Name:   state.Name,
					Result: outcome,
				},
			},
		}
		if err := stream.Send(&result); err != nil {
			return err
		}

		// TODO really stream this
		data := server.ApplyResult{
			Msg: &server.ApplyResult_Output{
				Output: &server.ApplyResultOutput{
					Output: state.Output.Bytes(),
				},
			},
		}
		if err := stream.Send(&data); err != nil {
			return err
		}
	}
	return nil
}

// Plan ...
func (h *HPTServer) Plan(ctx context.Context, conf *server.Config) (*server.PlanResult, error) {
	return nil, errors.New("unimplemented")
}

func whatHappened(state State) server.ApplyResultMetadata_Outcome {
	var m map[State]server.ApplyResultMetadata_Outcome
	m[Changed] = server.ApplyResultMetadata_CHANGED
	m[Unchanged] = server.ApplyResultMetadata_UNCHANGED
	return m[state]
}

type StormKeystore struct {
	DB *storm.DB
}

func (sk *StormKeystore) Allowed(pubkey curvetls.Pubkey) bool {
	var kp KeyPair
	// pass the value we are querying for as the second param
	if err := sk.DB.One("Pub", pubkey.String(), &kp); err == nil {
		return true
	}
	return false
}

// KeyPair is a database type that represents curvetls key pairs. A KeyPair
// must be in the database for each pure grpc client that wants to connect.
// Not used for grpc websocket clients.
type KeyPair struct {
	Name string `storm:"unique,id"`
	// Pub and Priv are base64 strings that represent curvetls keys
	// for servers or clients.
	Pub  string
	Priv string
}

// RetrieveServerKeys gets the curvetls public key for our hpt instance.
// Clients will need to know this server's public key in advance.
func RetrieveServerKeys(db *storm.DB) (curvetls.Pubkey, curvetls.Privkey, error) {
	var kp KeyPair
	err := db.One("Name", "server", &kp)
	if err != nil {
		return curvetls.Pubkey{}, curvetls.Privkey{}, err
	}
	priv, err := curvetls.PrivkeyFromString(kp.Priv)
	if err != nil {
		return curvetls.Pubkey{}, curvetls.Privkey{}, err
	}
	pub, err := curvetls.PubkeyFromString(kp.Pub)
	if err != nil {
		return curvetls.Pubkey{}, curvetls.Privkey{}, err
	}
	return pub, priv, nil
}
