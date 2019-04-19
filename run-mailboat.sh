#!/bin/bash

NPROC=$1
NMSG=100000
N=$(($NMSG * $NPROC))
TIMEFORMAT='real %R nuser %U sys %S (s)'
echo "== gomail $NPROC $NMSG $N `date` == "
for i in `seq 1 $NPROC`;
do
    echo "== gomail $i $((N / i))"
    ( GOMAIL_NPROC=$i GOMAIL_NITER=$((N / i)) go test -v -run Mixed)
    sleep 1
done    

