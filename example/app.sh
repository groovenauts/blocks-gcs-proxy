#!/bin/bash

usage() {
  echo usage: `basename $0` download_filepath downloads_dir uploads_dir suffix
}

if [ $# -ne 4 ]; then
  echo `basename $0`: missing operand 1>&2
  usage
  exit 1
fi

fpath=$1
downloads_dir=$2
uploads_dir=$3
suffix=$4

dl_localpath=${fpath##$downloads_dir/}
dl_bucket=$(echo $dl_localpath | cut -d '/' -f1)
dl_relpath=${dl_localpath##$dl_bucket/}
dl_fname=${dl_relpath##*/}
dl_dir=${dl_relpath%$dl_fname}

fname="${fpath##*/}"
fbase="${fname%.*}"
fext="${fpath##*.}"

ul_dir="$3/${dl_bucket}/${dl_dir}"
ul_path="${ul_dir}${fbase}-${suffix}.${fext}"

mkdir -p $ul_dir
cp $fpath $ul_path
