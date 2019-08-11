#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$DIR/.." || exit

NPROC=$1
NMSG=100000

if [ -z "$NPROC" ]; then
    echo "Usage $0 <nproc>" 1>&2
    exit 1
fi

N=$((NMSG * NPROC))
TIMEFORMAT='real %R nuser %U sys %S (s)'
go build -o /tmp/mailboat-bench ./cmd/mailboat-bench/main.go
echo "== mailboat $NPROC $NMSG $N $(date) == "
for i in $(seq 1 "$NPROC")
do
    echo "== mailboat $i $((N / i))"
    time ( GOMAIL_NPROC=$i GOMAIL_NITER=$((N / i)) /tmp/mailboat-bench)
    sleep 1
done
rm /tmp/mailboat-bench
