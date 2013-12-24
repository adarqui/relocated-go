#!/bin/bash

mkdir -p /tmp/{incoming,dest}

cp scripts/relocate.sh /tmp/

chmod 755 /tmp/relocate.sh

for i in `seq 1 50`; do touch /tmp/incoming/$i.mov; done
for i in `seq 1 20`; do touch /tmp/incoming/$i.wmv; done
for i in `seq 1 100`; do touch /tmp/incoming/$i.jpg; done
