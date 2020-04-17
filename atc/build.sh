#!/bin/sh

clang -O0 -g -DDEBUG -Wall -Wextra -Wpedantic -Wshadow -march=native -fPIE -fno-color-diagnostics -c main.c -o /tmp/atc.o \
    && clang -O0 -g -DDEBUG /tmp/atc.o -o /home/_/.local/bin/atc
