package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/Rudd-O/curvetls"
	"github.com/anxiousmodernman/hpt/proto/server"
	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var _ server.HPTServer = (*HPTServer)(nil)

// HPTServer is our implementation of the generated server.HPTServer interface.
// See the proto directory for the iterface definition in protobuf.
type HPTServer struct{}

// NewHPTServer is the constructor for our implementation of the generated gRPC
// interface.
func NewHPTServer(path string) (*HPTServer, *grpc.Server, error) {

	// We expect the db to exist. During initial setup, an authorized client
	// will create this DB for us and populate it with authorized client keys,
	// as well as our own key. We must have all authorized clients in the db,
	// because we cannot tolerate even accepting connections from unauthorized
	// clients.
	db, err := storm.Open(path)
	if err != nil {
		return nil, nil, err
	}

	// The curvetls interface from our fork.
	var keystore curvetls.KeyStore
	keystore = &StormKeystore{db}
	// curvetls transport security
	serverPub, serverPriv, err := RetrieveServerKeys(db)
	if err != nil {
		return nil, nil, fmt.Errorf("could not retrieve server's keypair: %v", err)
	}
	creds := curvetls.NewGRPCServerCredentials(serverPub, serverPriv, keystore)

	svr := grpc.NewServer(
		grpc.Creds(creds),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  30 * time.Second,
			Timeout:               10 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 30 * time.Second,
			MaxConnectionIdle:     30 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}))
	var h HPTServer
	server.RegisterHPTServer(svr, &h)
	return &h, svr, nil
}

// Apply accepts a config and provisions a target.
func (h *HPTServer) Apply(conf *server.Config, stream server.HPT_ApplyServer) error {

	c, err := NewConfigFromBytes(conf.Data)
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
			fmt.Println("done!")
			stream.Context().Done()
			return nil
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
	var m = make(map[State]server.ApplyResultMetadata_Outcome)
	m[Changed] = server.ApplyResultMetadata_CHANGED
	m[Unchanged] = server.ApplyResultMetadata_UNCHANGED
	return m[state]
}

// StormKeystore implements grpc.Crendentials.
type StormKeystore struct {
	DB *storm.DB
}

// Allowed is a method that implements a grpc.Credentials.
func (sk *StormKeystore) Allowed(pubkey curvetls.Pubkey) bool {
	var kp KeyPair
	// pass the value we are querying for as the second param
	log.Println("client pubkey", pubkey.String())
	err := sk.DB.One("Pub", pubkey.String(), &kp)
	if err == nil {
		return true
	}
	log.Println("Keystore error:", err)
	return false
}

// KeyPair is a database type that represents curvetls key pairs. A KeyPair
// must be in the database for each pure grpc client that wants to connect.
type KeyPair struct {
	Name string `storm:"unique,id"`
	// Pub and Priv are base64 strings for curvetls
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

// GenerateTargetServerKeys creates a boltdb keystore with a new
// targets server keys prepopulated. When a target runs an hpt
// agent with `hpt serve --keystore=/etc/hpt/keystore.db`, it is to
// make use of a keystore database generated by this function.
func GenerateTargetServerKeys(targetName, localKeystorePath string) ([]byte, error) {
	if targetName == "" {
		return nil, errors.New("must provide a target name")
	}

	priv, pub, err := curvetls.GenKeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "could not generate key pair")
	}
	serverKP := KeyPair{"server", pub.String(), priv.String()}
	// We don't need the private key in our local copy. It will only
	// live on the server.
	localKP := KeyPair{targetName, pub.String(), ""}

	db, err := storm.Open(localKeystorePath)
	if err != nil {
		return nil, errors.Wrap(err,
			fmt.Sprintf("error opening database at path: %s", localKeystorePath))
	}
	var clientKP KeyPair
	err = db.One("Name", "client", &clientKP)
	if err != nil {
		return nil, err
	}
	clientKP.Priv = "" // very important that we do this

	// TODO: We're overwriting targets by name right now. Probably want to
	// warn instead.
	if err := db.Save(&localKP); err != nil {
		return nil, err
	}
	defer db.Close()

	f, err := ioutil.TempFile("", "hpt_target")
	if err != nil {
		return nil, err
	}
	f.Close()
	outDB, err := storm.Open(f.Name())
	if err != nil {
		return nil, err
	}
	if err := outDB.Save(&serverKP); err != nil {
		return nil, err
	}
	if err := outDB.Save(&clientKP); err != nil {
		return nil, err
	}
	if err := outDB.Close(); err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ListTargetServerKeys is a utility method to print the named public keys
// in our keystore database.
func ListTargetServerKeys(localKeystorePath string) error {

	db, err := storm.Open(localKeystorePath)
	if err != nil {
		return err
	}
	var keypairs []KeyPair
	err = db.All(&keypairs)
	if err != nil {
		return err
	}
	for _, kp := range keypairs {
		fmt.Println("name\t:", kp.Name)
		fmt.Println("pub key\t:", kp.Pub)
	}
	return nil
}

// GenerateClientKeys adds/overwrites the "client" keypair in our local
// keystore. This should be done once, before target keys are generated,
// because targets need to know the client's public key. If the client's key
// changes, target keystores will need to be regenerated.
//
// TODO add a simpler key-rolling mechanism to transactionally roll keys
// across targets.
func GenerateClientKeys(localKeystorePath string) error {

	db, err := storm.Open(localKeystorePath)
	if err != nil {
		return err
	}
	priv, pub, err := curvetls.GenKeyPair()
	if err != nil {
		return errors.Wrap(err, "could not generate key pair")
	}
	// "client" is a magic name that identifies our client keypair. Since
	// Name must be unique, only one client keypair can be active at a time.
	kp := KeyPair{"client", pub.String(), priv.String()}
	return db.Save(&kp)
}
