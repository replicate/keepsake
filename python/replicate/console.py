import sys

from ._vendor.colors import color

# Parallel of cli/pkg/console/


def info(s):
    # TODO: word wrapping

    prompt = "═══╡ "
    continuation_prompt = "   │ "

    # TODO: explicit disabling of colors, respect NO_COLOR, etc
    if sys.stderr.isatty():
        prompt = color(prompt, style="faint")
        continuation_prompt = color(continuation_prompt, style="faint")

    for i, line in enumerate(s.split("\n")):
        if i == 0:
            print(prompt + line, file=sys.stderr)
        else:
            print(continuation_prompt + line, file=sys.stderr)
