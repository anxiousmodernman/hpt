package main

import (
	"context"

	"github.com/BurntSushi/toml"
	"github.com/anxiousmodernman/hpt/proto/server"
	"github.com/pkg/errors"
)

var _ server.HPTServer = (*HPTServer)(nil)

// HPTServer ...
type HPTServer struct {
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
