#!/bin/bash

SELF=$(basename "$0")
HOOKS_DIR=$(dirname "$PWD"/"$0")
GIT_COMMON_DIR=$(git rev-parse --git-common-dir) # finds the directory containing the `hooks` directory

for F in "$HOOKS_DIR"/*; do
    HOOK_NAME=$(basename "$F")
    if [ "$SELF" != "$HOOK_NAME" ] && [ -x "$F" ]; then
        echo "installing $F as $GIT_COMMON_DIR/hooks/$HOOK_NAME"
        ln -sf "$F" "$GIT_COMMON_DIR"/hooks/"$HOOK_NAME"
    fi
done
