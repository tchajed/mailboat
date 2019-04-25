#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$DIR/.." || exit

NPROC=$1
NMSG=100000

if [ -z "$NPROC" ]; then
    echo "Usage $0 <nproc>" 1>&2
    exit 1
fi

N=$(($NMSG * $NPROC))
TIMEFORMAT='real %R nuser %U sys %S (s)'
echo "== gomail $NPROC $NMSG $N `date` == "
for i in `seq 1 $NPROC`;
do
    echo "== gomail $i $((N / i))"
    ( GOMAIL_NPROC=$i GOMAIL_NITER=$((N / i)) go run ./cmd/mailboat-bench)
    sleep 1
done
