#!/bin/sh

for file_name in $(ls *.lb)
do
    # turns foo.lb into foo
    file_name_sans_ext=`basename $file_name .lb`
    echo "Build: $file_name_sans_ext ..."

    langboot $file_name_sans_ext.lb > $file_name_sans_ext.ll \
        && llc --relocation-model=pic $file_name_sans_ext.ll -o $file_name_sans_ext.s \
        && gcc $file_name_sans_ext.s -o ~/.local/bin/$file_name_sans_ext
done;
