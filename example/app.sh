#!/bin/bash

usage() {
  echo usage: `basename $0` src_path dest_path
}

if [ $# -ne 2 ]; then
  echo `basename $0`: missing operand 1>&2
  usage
  exit 1
fi

fpath=$1
fname="${fpath##*/}"
fbase="${fname%.*}"
fext="${fpath##*.}"
mkdir -p $2

cp $1 $2/${fbase}-$(date '+%Y%m%d-%T' | tr -d :).${fext}
