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

echo "dl_localpath=${dl_localpath}"
echo "dl_bucket=${dl_bucket}"
echo "dl_relpath=${dl_relpath}"
echo "dl_fname=${dl_fname}"
echo "dl_dir=${dl_dir}"


fname="${fpath##*/}"
fbase="${fname%.*}"
fext="${fpath##*.}"

echo "fname=${fname}"
echo "fbase=${fbase}"
echo "fext=${fext}"


ul_dir="$3/${dl_bucket}/${dl_dir}"
ul_path="${ul_dir}${fbase}-${suffix}.${fext}"

echo "ul_dir=${ul_dir}"
echo "ul_path=${ul_path}"

mkdir -p $ul_dir
cp $fpath $ul_path
