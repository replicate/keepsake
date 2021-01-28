# TODO: docstring
# TODO: rename to shared?

import functools
import tempfile
import os
from typing import Optional, Dict, Any, List
import subprocess
import atexit
import sys
import threading

import grpc  # type: ignore
from google.rpc import status_pb2, error_details_pb2  # type: ignore

from .servicepb.keepsake_pb2_grpc import DaemonStub
from .servicepb import keepsake_pb2 as pb
from . import pb_convert
from .experiment import Experiment
from .checkpoint import Checkpoint, PrimaryMetric
from . import exceptions
from . import console

# TODO(andreas): rename to keepsake-daemon
DAEMON_BINARY = os.path.join(os.path.dirname(__file__), "bin/keepsake-shared")


def handle_error(f):
    @functools.wraps(f)
    def wrapped(*args, **kwargs):
        try:
            return f(*args, **kwargs)
        except grpc.RpcError as e:
            code, name = e.code().value
            details = e.details()
            if name == "internal":
                status_code = get_status_code(e, details)
                if status_code:
                    raise handle_exception(status_code, details)
            raise Exception(details)

    return wrapped


def handle_exception(code, details):
    if code == "DOES_NOT_EXIST":
        return exceptions.DoesNotExist(details)
    if code == "READ_ERROR":
        return exceptions.ReadError(details)
    if code == "WRITE_ERROR":
        return exceptions.WriteError(details)
    if code == "REPOSITORY_CONFIGURATION_ERROR":
        return exceptions.RepositoryConfigurationError(details)
    if code == "INCOMPATIBLE_REPOSITORY_VERSION":
        return exceptions.IncompatibleRepositoryVersion(details)
    if code == "CORRUPTED_REPOSITORY_SPEC":
        return exceptions.CorruptedRepositorySpec(details)
    if code == "CONFIG_NOT_FOUND":
        return exceptions.ConfigNotFound(details)


def get_status_code(e, details):
    metadata = e.trailing_metadata()
    status_md = [x for x in metadata if is_status_detail(x)]
    if status_md:
        for md in status_md:
            st = status_pb2.Status()
            st.MergeFromString(md.value)
        if st.details:
            val = error_details_pb2.ErrorInfo()
            st.details[0].Unpack(val)
            return val.reason
    return None


def is_status_detail(x):
    return hasattr(x, "key") and x.key == "grpc-status-details-bin"


class Daemon:
    def __init__(self, project, socket_path=None, debug=False):
        self.project = project

        if socket_path is None:
            # create a new temporary file just to get a free name.
            # the Go GRPC server will create the file.
            f = tempfile.NamedTemporaryFile(
                prefix="keepsake-daemon-", suffix=".sock", delete=False
            )
            self.socket_path = f.name
            f.close()
        else:
            self.socket_path = socket_path

        # the Go GRPC server will fail to start if the socket file
        # already exists.
        os.unlink(self.socket_path)

        cmd = [DAEMON_BINARY]
        if self.project.repository:
            cmd += ["-R", self.project.repository]
        if self.project.directory:
            cmd += ["-D", self.project.directory]
        if debug:
            cmd += ["-v"]
        cmd.append(self.socket_path)
        self.process = subprocess.Popen(
            cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE
        )

        # need to wrap stdout and stderr for this to work in jupyter
        # notebooks. jupyter redefines sys.std{out,err} as custom
        # writers that eventually write the output to the notebook.
        self.stdout_thread = start_wrapped_pipe(self.process.stdout, sys.stdout)
        self.stderr_thread = start_wrapped_pipe(self.process.stderr, sys.stderr)

        atexit.register(self.cleanup)
        self.channel = grpc.insecure_channel("unix://" + self.socket_path)
        self.stub = DaemonStub(self.channel)

        TIMEOUT_SEC = 15
        grpc.channel_ready_future(self.channel).result(timeout=TIMEOUT_SEC)

        # TODO(andreas): catch daemon dying (bubble up an exception so we can fail on experiment.init())

    def cleanup(self):
        if self.process.poll() is None:  # check if process is still running:
            # the sigterm handler in the daemon process waits for any in-progress uploads etc. to finish.
            # the sigterm handler also deletes the socket file
            self.process.terminate()
            self.process.wait()

            # need to join these threads to avoid "could not acquire lock" error
            self.stdout_thread.join()
            self.stderr_thread.join()
        self.channel.close()

    @handle_error
    def create_experiment(
        self,
        path: Optional[str],
        params: Optional[Dict[str, Any]],
        command: Optional[str],
        python_packages: Dict[str, str],
        python_version: str,
        quiet: bool,
        disable_hearbeat: bool,
    ) -> Experiment:
        pb_experiment = pb.Experiment(
            params=pb_convert.value_map_to_pb(params),
            path=path,
            command=command,
            pythonPackages=python_packages,
            pythonVersion=python_version,
        )
        ret = self.stub.CreateExperiment(
            pb.CreateExperimentRequest(
                experiment=pb_experiment,
                disableHeartbeat=disable_hearbeat,
                quiet=quiet,
            ),
        )
        return pb_convert.experiment_from_pb(self.project, ret.experiment)

    @handle_error
    def create_checkpoint(
        self,
        experiment: Experiment,
        path: Optional[str],
        step: Optional[int],
        metrics: Optional[Dict[str, Any]],
        primary_metric: Optional[PrimaryMetric],
        quiet: bool,
    ) -> Checkpoint:
        pb_primary_metric = pb_convert.primary_metric_to_pb(primary_metric)
        pb_checkpoint = pb.Checkpoint(
            metrics=pb_convert.value_map_to_pb(metrics),
            path=path,
            primaryMetric=pb_primary_metric,
            step=step,
        )
        ret = self.stub.CreateCheckpoint(
            pb.CreateCheckpointRequest(checkpoint=pb_checkpoint, quiet=quiet)
        )
        return pb_convert.checkpoint_from_pb(experiment, ret.checkpoint)

    @handle_error
    def save_experiment(
        self, experiment: Experiment, quiet: bool,
    ):
        pb_experiment = pb_convert.experiment_to_pb(experiment)
        return self.stub.SaveExperiment(
            pb.SaveExperimentRequest(experiment=pb_experiment, quiet=quiet)
        )

    @handle_error
    def stop_experiment(self, experiment_id: str):
        self.stub.StopExperiment(pb.StopExperimentRequest(experimentID=experiment_id))

    @handle_error
    def get_experiment(self, experiment_id_prefix: str) -> Experiment:
        ret = self.stub.GetExperiment(
            pb.GetExperimentRequest(experimentIDPrefix=experiment_id_prefix),
        )
        return pb_convert.experiment_from_pb(self.project, ret.experiment)

    @handle_error
    def list_experiments(self) -> List[Experiment]:
        ret = self.stub.ListExperiments(pb.ListExperimentsRequest())
        return pb_convert.experiments_from_pb(self.project, ret.experiments)

    @handle_error
    def delete_experiment(self, experiment_id: str):
        self.stub.DeleteExperiment(
            pb.DeleteExperimentRequest(experimentID=experiment_id)
        )

    @handle_error
    def checkout_checkpoint(
        self, checkpoint_id_prefix: str, output_directory: str, quiet: bool
    ):
        self.stub.CheckoutCheckpoint(
            pb.CheckoutCheckpointRequest(
                checkpointIDPrefix=checkpoint_id_prefix,
                outputDirectory=output_directory,
                quiet=quiet,
            ),
        )

    @handle_error
    def experiment_is_running(self, experiment_id: str) -> str:
        ret = self.stub.GetExperimentStatus(
            pb.GetExperimentStatusRequest(experimentID=experiment_id)
        )
        return ret.status == pb.GetExperimentStatusReply.Status.RUNNING


def start_wrapped_pipe(pipe, writer):
    def wrap_pipe(pipe, writer):
        with pipe:
            for line in iter(pipe.readline, b""):
                writer.write(line)
                writer.flush()

    # if writer is normal sys.std{out,err}, it can't
    # write bytes directly.
    # see https://stackoverflow.com/a/908440/135797
    if hasattr(writer, "buffer"):
        writer = writer.buffer

    thread = threading.Thread(target=wrap_pipe, args=[pipe, writer], daemon=True)
    thread.start()
    return thread
