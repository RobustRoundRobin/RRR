#!/bin/sh
cd $(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)/.. && exec $(which python3) -m tools.gcpblob "$@"


