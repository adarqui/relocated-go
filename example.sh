#!/bin/bash

mkdir -p /tmp/company1/{movies,movies2,dest,bin}
mkdir -p /tmp/company2/{misc,out,bin,new_releases}

cp scripts/relocate.sh /tmp/company1/bin
cp scripts/relocate.sh /tmp/company2/bin

chmod 755 /tmp/company*/bin/*

for i in `seq 1 50`; do touch /tmp/company1/movies/$i.mov; done
for i in `seq 1 20`; do touch /tmp/company1/movies2/$i.wmv; done
for i in `seq 1 100`; do touch /tmp/company2/new_releases/$i.jpg; done
