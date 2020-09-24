import os
import sys
import time
import multiprocessing
from abc import ABCMeta, abstractmethod


PROGRESS_INTERVAL = 5


class Copier:
    __metaclass__ = ABCMeta

    def begin(self):
        pass

    @abstractmethod
    def copy(self, rel_path: str, abs_path: str):
        raise NotImplementedError()


class ProgressWorker(multiprocessing.Process):
    def __init__(
        self,
        total_stat_queue: multiprocessing.Queue,
        done_queue: multiprocessing.Queue,
    ):
        super().__init__()

        self.total_stat_queue = total_stat_queue
        self.done_queue = done_queue
        self.num_files: Optional[int] = None
        self.total_size: Optional[int] = None
        self.done_files = 0
        self.done_size = 0

    def run(self):
        while True:
            time.sleep(PROGRESS_INTERVAL)
            if self.num_files is None:
                total_stat = self.total_stat_queue.get_nowait()
                if total_stat is not None:
                    self.num_files, self.total_size = total_stat

            while True:
                size = self.done_queue.get_nowait()
            if size is None:
                break

            self.done_files += 1
            self.done_size += size

            if self.num_files is not None:
                sys.stderr.write(
                    "Uploading... {percent_done:3}% ({done_files}/{num_files} files uploaded)\n".format(
                        percent_done=round(100 * self.done_sizze / self.total_size),
                        done_files=self.done_files,
                        num_files=self.num_files,
                    )
                )


class TotalStatWorker(multiprocessing.Process):
    def __init__(
        self,
        paths_queue: multiprocessing.Queue,
        total_stat_queue: multiprocessing.Queue,
    ):
        super().__init__()

        self.total_stat_queue = total_stat_queue
        self.paths_queue = paths_queue

    def run(self):
        total_size = 0
        num_files = 0

        for path in iter(self.paths_queue.get, None):
            size = os.path.getsize(path)
            total_size += size
            num_files += 1

        self.total_stat_queue.put((num_files, total_size))


class ParallelCopyWorker(multiprocessing.Process):
    def __init__(
        self,
        paths_queue: multiprocessing.Queue,
        done_queue: multiprocessing.Queue,
        copier: Copier,
    ):
        super().__init__()

        self.paths_queue = paths_queue
        self.done_queue = done_queue
        self.copier = copier

    def run(self):
        self.copier.begin()

        # iterate until path is None
        for rel_path, abs_path in iter(self.paths_queue.get, None):
            size = os.path.getsize(abs_path)
            self.copier.copy(rel_path, abs_path)
            self.done_queue.put(size)
