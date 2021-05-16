"""Utilities to promote consistent command line behaviour

Notice: Python 3.6 is the maximum python version for this package
"""
import subprocess as sp
from pathlib import Path
import sys
import traceback


class Error(Exception):
    pass


def print_exc():
    """Compact representation of current exception

    Single line tracebacks are individually a little confusing but prevent
    other useful output from being obscured"""

    exc_info = sys.exc_info()
    trace = [
        f"{Path(fn).name}[{ln}].{fun}:\"{txt}\""
        for (fn, ln, fun, txt) in traceback.extract_tb(exc_info[2])
    ] + [f"{exc_info[0].__name__}:{exc_info[1]}"]

    print("->".join(trace), file=sys.stderr)


def run_and_exit(runner, *errors):
    """
    Wrap a main entry point so that errors are caught and printed in a
    sensible way.

    Exception Handling:

    CalledProcessError is caught and, if available, stderr is printed to
    stderr. stdout is printed if availalbe.

    Errors: A tuple of all errors that should be reported simply by::

        print(repr(exc), str(exc), file=sys.stderr)

    Exception: is caught and a compact single line traceback is printed to
    stderr.

    If no exceptions occur, sys.exit(0) is called. Otherwise sys.exit(-1)
    """
    try:
        runner()
        sys.exit(0)
    except sp.CalledProcessError as cpe:
        if cpe.stdout:
            print(cpe.stdout)
        print(cpe.stderr, file=sys.stderr)
        sys.exit(cpe.returncode)

    except errors as exc:
        print(repr(exc), str(exc), file=sys.stderr)
    except KeyboardInterrupt:
        sys.exit(-1)
    except Exception:
        print_exc()

    sys.exit(-1)

