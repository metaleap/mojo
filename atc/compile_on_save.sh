#!/bin/sh

clang -O0 -g -DDEBUG -Wall -Wextra -Wpedantic -Wshadow -march=native -fPIE -fno-color-diagnostics -c main.c -o /tmp/atc.o \
    && clang -O0 -g -DDEBUG /tmp/atc.o -o $HOME/.local/bin/atc




# // note for later from https://embeddedartistry.com/blog/2017/07/05/printf-a-limited-number-of-characters-from-a-string/
# // Only 5 characters printed. When using %.*s, add a value before your string variable to specify the length.
# printf("Here are the first 5 characters: %.*s\n", 5, mystr); //5 here refers to # of characters
