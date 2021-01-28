package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/replicate/keepsake/go/pkg/config"
	"github.com/replicate/keepsake/go/pkg/param"
	"github.com/replicate/keepsake/go/pkg/project"
	"github.com/replicate/keepsake/go/pkg/servicepb"
)

func fullCheckpointPb() *servicepb.Checkpoint {
	return &servicepb.Checkpoint{
		Id:      "foo",
		Created: timestamppb.New(time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC)),
		Path:    ".",
		Step:    123,
		Metrics: map[string]*servicepb.ParamType{
			"myint":    {Value: &servicepb.ParamType_IntValue{IntValue: 456}},
			"myfloat":  {Value: &servicepb.ParamType_FloatValue{FloatValue: 7.89}},
			"mystring": {Value: &servicepb.ParamType_StringValue{StringValue: "value"}},
			"mytrue":   {Value: &servicepb.ParamType_BoolValue{BoolValue: true}},
			"myfalse":  {Value: &servicepb.ParamType_BoolValue{BoolValue: false}},
			"mylist":   {Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: "[1,2,3]"}},
			"mymap":    {Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: `{"bar":"baz"}`}},
		},
		PrimaryMetric: &servicepb.PrimaryMetric{
			Name: "myfloat",
			Goal: servicepb.PrimaryMetric_MAXIMIZE,
		},
	}
}

func fullCheckpoint() *project.Checkpoint {
	return &project.Checkpoint{
		ID:      "foo",
		Created: time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC),
		Path:    ".",
		Step:    123,
		Metrics: map[string]param.Value{
			"myint":    param.Int(456),
			"myfloat":  param.Float(7.89),
			"mystring": param.String("value"),
			"mytrue":   param.Bool(true),
			"myfalse":  param.Bool(false),
			"mylist":   param.Object([]interface{}{1.0, 2.0, 3.0}),
			"mymap":    param.Object(map[string]interface{}{"bar": "baz"}),
		},
		PrimaryMetric: &project.PrimaryMetric{Name: "myfloat", Goal: "maximize"},
	}
}

func emptyCheckpointPb() *servicepb.Checkpoint {
	return &servicepb.Checkpoint{
		Id:      "foo",
		Created: timestamppb.New(time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC)),
		Step:    0,
	}
}

func emptyCheckpoint() *project.Checkpoint {
	return &project.Checkpoint{
		ID:      "foo",
		Created: time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC),
		Step:    0,
	}
}

func fullExperimentPb() *servicepb.Experiment {
	t := time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC)
	return &servicepb.Experiment{
		Id:      "foo",
		Created: timestamppb.New(t),
		User:    "myuser",
		Host:    "myhost",
		Command: "mycmd",
		Config:  &servicepb.Config{Repository: "myrepo", Storage: ""},
		Path:    "mypath",
		Params: map[string]*servicepb.ParamType{
			"myint":    {Value: &servicepb.ParamType_IntValue{IntValue: 456}},
			"myfloat":  {Value: &servicepb.ParamType_FloatValue{FloatValue: 7.89}},
			"mystring": {Value: &servicepb.ParamType_StringValue{StringValue: "value"}},
			"mytrue":   {Value: &servicepb.ParamType_BoolValue{BoolValue: true}},
			"myfalse":  {Value: &servicepb.ParamType_BoolValue{BoolValue: false}},
			"mylist":   {Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: "[1,2,3]"}},
			"mymap":    {Value: &servicepb.ParamType_ObjectValueJson{ObjectValueJson: `{"bar":"baz"}`}},
		},
		PythonPackages:  map[string]string{"pkg1": "1.1", "pkg2": "2.2"},
		KeepsakeVersion: "1.2.3",
		Checkpoints: []*servicepb.Checkpoint{
			{
				Id:      "c1",
				Created: timestamppb.New(t.Add(time.Minute * 1)),
				Step:    1,
			},
			{
				Id:      "c2",
				Created: timestamppb.New(t.Add(time.Minute * 2)),
				Step:    2,
			},
		},
	}
}

func fullExperiment() *project.Experiment {
	t := time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC)
	return &project.Experiment{
		ID:      "foo",
		Created: t,
		User:    "myuser",
		Host:    "myhost",
		Command: "mycmd",
		Config:  &config.Config{Repository: "myrepo", Storage: ""},
		Path:    "mypath",
		Params: map[string]param.Value{
			"myint":    param.Int(456),
			"myfloat":  param.Float(7.89),
			"mystring": param.String("value"),
			"mytrue":   param.Bool(true),
			"myfalse":  param.Bool(false),
			"mylist":   param.Object([]interface{}{1.0, 2.0, 3.0}),
			"mymap":    param.Object(map[string]interface{}{"bar": "baz"}),
		},
		PythonPackages:  map[string]string{"pkg1": "1.1", "pkg2": "2.2"},
		KeepsakeVersion: "1.2.3",
		Checkpoints: []*project.Checkpoint{
			{ID: "c1", Created: t.Add(time.Minute * 1), Step: 1},
			{ID: "c2", Created: t.Add(time.Minute * 2), Step: 2},
		},
	}
}

func emptyExperimentPb() *servicepb.Experiment {
	return &servicepb.Experiment{
		Id:      "foo",
		Created: timestamppb.New(time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC)),
	}
}

func emptyExperiment() *project.Experiment {
	return &project.Experiment{
		ID:      "foo",
		Created: time.Date(2020, 12, 7, 1, 13, 29, 192682, time.UTC),
	}
}

func TestConvertCheckpointFromPb(t *testing.T) {
	chkPb := fullCheckpointPb()
	expected := fullCheckpoint()
	require.Equal(t, expected, checkpointFromPb(chkPb))
}

func TestConvertEmptyCheckpointFromPb(t *testing.T) {
	chkPb := emptyCheckpointPb()
	expected := emptyCheckpoint()
	require.Equal(t, expected, checkpointFromPb(chkPb))
}

func TestConvertExperimentFromPb(t *testing.T) {
	expPb := fullExperimentPb()
	expected := fullExperiment()
	require.Equal(t, expected, experimentFromPb(expPb))
}

func TestConvertEmptyExperimentFromPb(t *testing.T) {
	expPb := emptyExperimentPb()
	expected := emptyExperiment()
	require.Equal(t, expected, experimentFromPb(expPb))
}

func TestConvertCheckpointToPb(t *testing.T) {
	chk := fullCheckpoint()
	expected := fullCheckpointPb()
	require.Equal(t, expected, checkpointToPb(chk))
}

func TestConvertEmptyCheckpointToPb(t *testing.T) {
	chk := emptyCheckpoint()
	expected := emptyCheckpointPb()
	require.Equal(t, expected, checkpointToPb(chk))
}

func TestConvertExperimentToPb(t *testing.T) {
	exp := fullExperiment()
	expected := fullExperimentPb()
	require.Equal(t, expected, experimentToPb(exp))
}

func TestConvertEmptyExperimentToPb(t *testing.T) {
	exp := emptyExperiment()
	expected := emptyExperimentPb()
	require.Equal(t, expected, experimentToPb(exp))
}
