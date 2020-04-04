#!/bin/sh

for file_name in $(ls *.ll)
do
    # turns foo.ll into foo
    file_name_sans_ext=`basename $file_name .ll`

    llc --relocation-model=pic $file_name_sans_ext.ll -o $file_name_sans_ext.s
    gcc $file_name_sans_ext.s -o ~/.local/bin/$file_name_sans_ext
done;
