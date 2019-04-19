#!/bin/bash

NUSER=100

rm -rf /tmp/mailtest

mkdir /tmp/mailtest
mkdir /tmp/mailtest/spool

echo "Create $NUSER mailboxes"
for i in `seq 0 $NUSER`;
do
    mkdir /tmp/mailtest/user$i
done    

