#! /bin/bash

# ./gen-go-data.sh ~/Downloads/mydata4vipweek2.dat main

function usage() {
	echo "usage: $0 <ipdata.dat> <go package name>" >&2
	exit 1
}

test $# -eq 2 || usage

test -f $1 || { echo "$1 not exist." && exit 1; }

OUT=data.go

cat << EOF > $OUT
package $2

var data = []byte{
$(xxd -i $1 | grep ','),
}
EOF
