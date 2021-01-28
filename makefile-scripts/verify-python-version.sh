#!/bin/bash -eu

RECOMMENDED_PYTHON_VERSION="3.8.6"
INSTALL_MESSAGE="We recommend using pyenv to install Python 3 and pip.
Follow the instructions on https://github.com/pyenv/pyenv-installer to install pyenv.
Then install Python 3:
  $ pyenv install $RECOMMENDED_PYTHON_VERSION
  $ pyenv local $RECOMMENDED_PYTHON_VERSION
Or, if you prefer to set the Python version globally:
  $ pyenv global $RECOMMENDED_PYTHON_VERSION"

PYTHON_VERSION=$(python --version 2>&1 || true)

if $(echo "$PYTHON_VERSION" | grep -q -E 'python: (command )?not found'); then
    echo "ERROR: Python is not installed."
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

if $(echo "$PYTHON_VERSION" | grep -q -E -v "^Python "); then
    echo "ERROR: failed to determine Python version, 'python --version' returned:"
    echo "  $PYTHON_VERSION"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

PYTHON_VERSION_NUMBER=$(echo "$PYTHON_VERSION" | sed -E 's/^Python (.+)$/\1/')

if $(echo "$PYTHON_VERSION_NUMBER" | grep -q -E -v '^3\.[6789]'); then
    echo "ERROR: Unsupported Python version: $PYTHON_VERSION_NUMBER"
    echo "Keepsake requires Python >= 3.6.0"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

PIP_VERSION=$(pip --version 2>&1 || true)

if $(echo "$PIP_VERSION" | grep -q -E 'pip: (command )?not found'); then
    echo "ERROR: pip is not installed."
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

if $(echo "$PIP_VERSION" | grep -q -E -v "pip .*python"); then
    echo "ERROR: failed to determine pip version, 'pip --version' returned:"
    echo "  $PIP_VERSION"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi

PIP_PYTHON_VERSION_NUMBER=$(echo "$PIP_VERSION" | sed -E 's/^pip.*\(python (.+)\)$/\1/')

if $(echo "$PIP_PYTHON_VERSION_NUMBER" | grep -q -E -v '^3\.[6789]'); then
    echo "ERROR: Unsupported Python version in pip: $PIP_PYTHON_VERSION_NUMBER"
    echo "Keepsake requires pip with Python >= 3.6.0"
    echo
    echo "$INSTALL_MESSAGE"
    exit 1
fi
