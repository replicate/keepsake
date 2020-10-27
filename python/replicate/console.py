import enum
from functools import wraps
import sys

from ._vendor.colors import color

# Parallel of go/pkg/console/


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


# Replicate should never break your training
def catch_and_print_exceptions(msg=None, return_value=None):
    def decorator(f):
        @wraps(f)
        def wrapper(*args, **kwargs):
            try:
                return f(*args, **kwargs)
            except Exception as e:  # pylint: disable=broad-except
                if msg is not None:
                    error(f"{msg}: {str(e)}")
                else:
                    error(str(e))
                return return_value

        return wrapper

    return decorator
