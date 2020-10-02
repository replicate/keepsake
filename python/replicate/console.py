import enum
import sys

from ._vendor.colors import color

# Parallel of cli/pkg/console/


class Level(enum.Enum):
    INFO = "INFO"
    WARN = "WARN"
    ERROR = "ERROR"


def info(s: str):
    log(s, Level.INFO)


def warn(s: str):
    log(s, Level.WARN)


def error(s: str):
    log(s, Level.ERROR)


def log(s: str, level: Level):
    # TODO: word wrapping

    prompt = "═══╡ "
    continuation_prompt = "   │ "

    # TODO: explicit disabling of colors, respect NO_COLOR, etc
    if sys.stderr.isatty():
        kwargs = {"style": "faint"}
        if level == Level.WARN:
            kwargs = {"fg": "yellow"}
        elif level == Level.ERROR:
            kwargs = {"fg": "red"}

        prompt = color(prompt, **kwargs)
        continuation_prompt = color(continuation_prompt, **kwargs)

    for i, line in enumerate(s.split("\n")):
        if i == 0:
            print(prompt + line, file=sys.stderr)
        else:
            print(continuation_prompt + line, file=sys.stderr)
