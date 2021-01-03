package shared

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/param"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/servicepb"
)

// convert from protobuf

func checkpointsFromPb(checkpointsPb []*servicepb.Checkpoint) []*project.Checkpoint {
	if checkpointsPb == nil {
		return nil
	}
	ret := make([]*project.Checkpoint, len(checkpointsPb))
	for i, chkPb := range checkpointsPb {
		ret[i] = checkpointFromPb(chkPb)
	}
	return ret
}

func checkpointFromPb(chkPb *servicepb.Checkpoint) *project.Checkpoint {
	return &project.Checkpoint{
		ID:            chkPb.Id,
		Created:       chkPb.Created.AsTime(),
		Metrics:       valueMapFromPb(chkPb.Metrics),
		Step:          chkPb.Step,
		Path:          chkPb.Path,
		PrimaryMetric: primaryMetricFromPb(chkPb.PrimaryMetric),
	}
}

func experimentFromPb(expPb *servicepb.Experiment) *project.Experiment {
	return &project.Experiment{
		ID:               expPb.Id,
		Created:          expPb.Created.AsTime(),
		Params:           valueMapFromPb(expPb.Params),
		Host:             expPb.Host,
		User:             expPb.User,
		Config:           configFromPb(expPb.Config),
		Command:          expPb.Command,
		Path:             expPb.Path,
		PythonPackages:   expPb.PythonPackages,
		PythonVersion:    expPb.PythonVersion,
		Checkpoints:      checkpointsFromPb(expPb.Checkpoints),
		ReplicateVersion: expPb.ReplicateVersion,
	}
}

func configFromPb(confPb *servicepb.Config) *config.Config {
	var conf *config.Config
	if confPb != nil {
		conf = &config.Config{Repository: confPb.Repository, Storage: confPb.Storage}
	}
	return conf
}

func primaryMetricFromPb(pmPb *servicepb.PrimaryMetric) *project.PrimaryMetric {
	if pmPb == nil {
		return nil
	}
	var goal project.MetricGoal
	switch pmPb.Goal {
	case servicepb.PrimaryMetric_MAXIMIZE:
		goal = project.GoalMaximize
	case servicepb.PrimaryMetric_MINIMIZE:
		goal = project.GoalMinimize
	}
	return &project.PrimaryMetric{
		Name: pmPb.Name,
		Goal: goal,
	}
}

func valueMapFromPb(pb map[string]*servicepb.ParamType) map[string]param.Value {
	if len(pb) == 0 {
		return nil
	}

	params := map[string]param.Value{}
	for k, v := range pb {
		params[k] = valueFromPb(v)
	}
	return params
}

func valueFromPb(pb *servicepb.ParamType) param.Value {
	switch pb.Value.(type) {
	case *servicepb.ParamType_BoolValue:
		return param.Bool(pb.GetBoolValue())
	case *servicepb.ParamType_IntValue:
		return param.Int(pb.GetIntValue())
	case *servicepb.ParamType_FloatValue:
		return param.Float(pb.GetFloatValue())
	case *servicepb.ParamType_StringValue:
		return param.String(pb.GetStringValue())
	case *servicepb.ParamType_ObjectValueJson:
		return param.ParseFromString(pb.GetObjectValueJson())
	}
	panic(fmt.Sprintf("Unknown param type: %v", pb)) // should never happen
}

// convert to protobuf

func experimentsToPb(experiments []*project.Experiment) []*servicepb.Experiment {
	ret := make([]*servicepb.Experiment, len(experiments))
	for i, exp := range experiments {
		ret[i] = experimentToPb(exp)
	}
	return ret
}

func experimentToPb(exp *project.Experiment) *servicepb.Experiment {
	return &servicepb.Experiment{
		Id:               exp.ID,
		Created:          timestamppb.New(exp.Created),
		Params:           valueMapToPb(exp.Params),
		Host:             exp.Host,
		User:             exp.User,
		Config:           configToPb(exp.Config),
		Command:          exp.Command,
		Path:             exp.Path,
		PythonPackages:   exp.PythonPackages,
		PythonVersion:    exp.PythonVersion,
		ReplicateVersion: exp.ReplicateVersion,
		Checkpoints:      checkpointsToPb(exp.Checkpoints),
	}
}

func configToPb(conf *config.Config) *servicepb.Config {
	if conf == nil {
		return nil
	}
	return &servicepb.Config{
		Repository: conf.Repository,
		Storage:    conf.Storage, // deprecated
	}
}

func checkpointsToPb(checkpoints []*project.Checkpoint) []*servicepb.Checkpoint {
	if checkpoints == nil {
		return nil
	}
	ret := make([]*servicepb.Checkpoint, len(checkpoints))
	for i, chk := range checkpoints {
		ret[i] = checkpointToPb(chk)
	}
	return ret
}

func checkpointToPb(chk *project.Checkpoint) *servicepb.Checkpoint {
	if chk == nil {
		return nil
	}
	return &servicepb.Checkpoint{
		Id:            chk.ID,
		Created:       timestamppb.New(chk.Created),
		Step:          chk.Step,
		Metrics:       valueMapToPb(chk.Metrics),
		Path:          chk.Path,
		PrimaryMetric: primaryMetricToPb(chk.PrimaryMetric),
	}
}

func primaryMetricToPb(pm *project.PrimaryMetric) *servicepb.PrimaryMetric {
	var pbPrimaryMetric *servicepb.PrimaryMetric
	if pm != nil {
		var goal servicepb.PrimaryMetric_Goal
		switch pm.Goal {
		case project.GoalMaximize:
			goal = servicepb.PrimaryMetric_MAXIMIZE
		case project.GoalMinimize:
			goal = servicepb.PrimaryMetric_MINIMIZE
		}

		pbPrimaryMetric = &servicepb.PrimaryMetric{
			Name: pm.Name,
			Goal: goal,
		}
	}
	return pbPrimaryMetric
}

func valueMapToPb(m map[string]param.Value) map[string]*servicepb.ParamType {
	if len(m) == 0 {
		return nil
	}

	pbMap := map[string]*servicepb.ParamType{}
	for k, v := range m {
		pbMap[k] = valueToPb(v)
	}
	return pbMap
}

func valueToPb(v param.Value) *servicepb.ParamType {
	switch v.Type() {
	case param.TypeBool:
		return &servicepb.ParamType{Value: &servicepb.ParamType_BoolValue{BoolValue: v.BoolVal()}}
	case param.TypeInt:
		return &servicepb.ParamType{Value: &servicepb.ParamType_IntValue{IntValue: v.IntVal()}}
	case param.TypeFloat:
		return &servicepb.ParamType{Value: &servicepb.ParamType_FloatValue{FloatValue: v.FloatVal()}}
	case param.TypeString:
		return &servicepb.ParamType{Value: &servicepb.ParamType_StringValue{StringValue: v.StringVal()}}
	case param.TypeObject:
		return &servicepb.ParamType{Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: v.String()}}
	case param.TypeNone:
		return &servicepb.ParamType{Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: "null"}}
	}
	panic("Uninitiazlied param.Value") // should never happen
}
