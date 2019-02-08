#! /usr/bin/env bash

go build -o gson;
./gson -collatesort example1 > out1
go build -tags n1ql -o gsonn1ql
./gsonn1ql -n1qlsort example1 > out2
rm out1 out2
