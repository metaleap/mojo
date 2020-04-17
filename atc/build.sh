#!/bin/sh

clang -O0 -g -Wall -Wextra -Wpedantic -Wshadow -march=native -fPIE -c main.c -o /tmp/atc.o && clang -O0 -g /tmp/atc.o -o /home/_/.local/bin/atc
