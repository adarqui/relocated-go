#!/bin/bash

export BIGNUM=1000

for i in `seq 51 100`; do

	for j in `seq 1 $(($RANDOM % 1000))|tail -n 1`; do head -c $(($j*$BIGNUM)) /dev/zero > $i.mov; done

done
