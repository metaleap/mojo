#!/bin/sh

llc --relocation-model=pic atb.ll -o atb.s
gcc atb.s -o ~/.local/bin/atb
