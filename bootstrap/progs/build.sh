#!/bin/sh

for file_name in $(ls app_*.at)
do
    # turns foo.at into foo
    file_name_sans_ext=`basename $file_name .at`
    echo "Build: $file_name_sans_ext ..."

    cat std_*.at $file_name_sans_ext.at | bootstrap > $file_name_sans_ext.ll \
        && llc --relocation-model=pic $file_name_sans_ext.ll -o $file_name_sans_ext.s \
        && clang $file_name_sans_ext.s -o ~/.local/bin/$file_name_sans_ext
done;
