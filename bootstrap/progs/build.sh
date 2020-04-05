#!/bin/sh

for file_name in $(ls *.at)
do
    # turns foo.at into foo
    file_name_sans_ext=`basename $file_name .at`
    echo "Build: $file_name_sans_ext ..."

    bootstrap $file_name_sans_ext.at > $file_name_sans_ext.ll \
        && llc --relocation-model=pic $file_name_sans_ext.ll -o $file_name_sans_ext.s \
        && gcc $file_name_sans_ext.s -o ~/.local/bin/$file_name_sans_ext
done;
