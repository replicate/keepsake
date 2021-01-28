import datetime

from keepsake import pb_convert
from keepsake.checkpoint import Checkpoint, PrimaryMetric, CheckpointList
from keepsake.experiment import Experiment
from keepsake.servicepb import keepsake_pb2 as pb
from keepsake.project import Project


def full_checkpoint_pb():
    return pb.Checkpoint(
        id="foo",
        created=pb_convert.timestamp_to_pb(
            datetime.datetime(2020, 12, 7, 1, 13, 29, 192682)
        ),
        path=".",
        step=123,
        metrics={
            "myint": pb.ParamType(intValue=456),
            "myfloat": pb.ParamType(floatValue=7.89),
            "mystring": pb.ParamType(stringValue="value"),
            "mytrue": pb.ParamType(boolValue=True),
            "myfalse": pb.ParamType(boolValue=False),
            "mylist": pb.ParamType(objectValueJson="[1, 2, 3]"),
            "mymap": pb.ParamType(objectValueJson='{"bar": "baz"}'),
        },
        primaryMetric=pb.PrimaryMetric(
            name="myfloat", goal=pb.PrimaryMetric.Goal.MAXIMIZE
        ),
    )


def full_checkpoint():
    return Checkpoint(
        id="foo",
        created=datetime.datetime(2020, 12, 7, 1, 13, 29, 192682),
        path=".",
        step=123,
        metrics={
            "myint": 456,
            "myfloat": 7.89,
            "mystring": "value",
            "mytrue": True,
            "myfalse": False,
            "mylist": [1, 2, 3],
            "mymap": {"bar": "baz"},
        },
        primary_metric=PrimaryMetric(name="myfloat", goal="maximize"),
    )


def empty_checkpoint_pb():
    return pb.Checkpoint(
        id="foo",
        created=pb_convert.timestamp_to_pb(
            datetime.datetime(2020, 12, 7, 1, 13, 29, 192682)
        ),
        step=0,
    )


def empty_checkpoint():
    return Checkpoint(
        id="foo", created=datetime.datetime(2020, 12, 7, 1, 13, 29, 192682), step=0,
    )


def full_experiment_pb():
    t = datetime.datetime(2020, 12, 7, 1, 13, 29, 192682)
    return pb.Experiment(
        id="foo",
        created=pb_convert.timestamp_to_pb(t),
        user="myuser",
        host="myhost",
        command="mycmd",
        config=pb.Config(repository="myrepo", storage=""),
        path="mypath",
        params={
            "myint": pb.ParamType(intValue=456),
            "myfloat": pb.ParamType(floatValue=7.89),
            "mystring": pb.ParamType(stringValue="value"),
            "mytrue": pb.ParamType(boolValue=True),
            "myfalse": pb.ParamType(boolValue=False),
            "mylist": pb.ParamType(objectValueJson="[1, 2, 3]"),
            "mymap": pb.ParamType(objectValueJson='{"bar": "baz"}'),
        },
        pythonPackages={"pkg1": "1.1", "pkg2": "2.2"},
        keepsakeVersion="1.2.3",
        checkpoints=[
            pb.Checkpoint(
                id="c1",
                created=pb_convert.timestamp_to_pb(t + datetime.timedelta(minutes=1)),
                step=1,
            ),
            pb.Checkpoint(
                id="c2",
                created=pb_convert.timestamp_to_pb(t + datetime.timedelta(minutes=2)),
                step=2,
            ),
        ],
    )


def full_experiment(project):
    t = datetime.datetime(2020, 12, 7, 1, 13, 29, 192682)
    return Experiment(
        project=project,
        id="foo",
        created=t,
        user="myuser",
        host="myhost",
        command="mycmd",
        config={"repository": "myrepo", "storage": ""},
        path="mypath",
        params={
            "myint": 456,
            "myfloat": 7.89,
            "mystring": "value",
            "mytrue": True,
            "myfalse": False,
            "mylist": [1, 2, 3],
            "mymap": {"bar": "baz"},
        },
        python_packages={"pkg1": "1.1", "pkg2": "2.2"},
        keepsake_version="1.2.3",
        checkpoints=CheckpointList(
            [
                Checkpoint(id="c1", created=t + datetime.timedelta(minutes=1), step=1,),
                Checkpoint(id="c2", created=t + datetime.timedelta(minutes=2), step=2,),
            ]
        ),
    )


def empty_experiment_pb():
    return pb.Experiment(
        id="foo",
        created=pb_convert.timestamp_to_pb(
            datetime.datetime(2020, 12, 7, 1, 13, 29, 192682)
        ),
    )


def empty_experiment(project):
    return Experiment(
        project=project,
        id="foo",
        created=datetime.datetime(2020, 12, 7, 1, 13, 29, 192682),
    )


def test_checkpoint_from_pb():
    chk_pb = full_checkpoint_pb()
    expected = full_checkpoint()
    assert pb_convert.checkpoint_from_pb(None, chk_pb) == expected


def test_empty_checkpoint_from_pb():
    chk_pb = empty_checkpoint_pb()
    expected = empty_checkpoint()
    assert pb_convert.checkpoint_from_pb(None, chk_pb) == expected


def test_experiment_from_pb():
    exp_pb = full_experiment_pb()
    project = Project()
    expected = full_experiment(project)
    assert pb_convert.experiment_from_pb(project, exp_pb) == expected


def test_empty_experiment_from_pb():
    exp_pb = empty_experiment_pb()
    project = Project()
    expected = empty_experiment(project)
    assert pb_convert.experiment_from_pb(project, exp_pb) == expected


def test_checkpoint_to_pb():
    chk = full_checkpoint()
    expected = full_checkpoint_pb()
    assert pb_convert.checkpoint_to_pb(chk) == expected


def test_empty_checkpoint_to_pb():
    chk = empty_checkpoint()
    expected = empty_checkpoint_pb()
    assert pb_convert.checkpoint_to_pb(chk) == expected


def test_experiment_to_pb():
    project = Project()
    exp = full_experiment(project)
    expected = full_experiment_pb()
    assert pb_convert.experiment_to_pb(exp) == expected


def test_empty_experiment_to_pb():
    project = Project()
    exp = empty_experiment(project)
    expected = empty_experiment_pb()
    assert pb_convert.experiment_to_pb(exp) == expected
