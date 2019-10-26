# Mailboat: a verified mail server

[![Build Status](https://travis-ci.com/tchajed/mailboat.svg?branch=master)](https://travis-ci.com/tchajed/mailboat)

Mailboat is a qmail-like mail server with a proof in [Perennial](https://github.com/mit-pdos/perennial). The proof shows that delivering, reading, and deleting mail are atomic with respect to other threads and crashes, and that operations are durable as soon as they return.

## Benchmarking

Run `./scripts/run-mailboat.sh` to run a mixed-workload benchmark, which is output from the test `TestMixedLoad`.
