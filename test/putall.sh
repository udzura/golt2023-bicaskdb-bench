#!/bin/bash

for f in `ls test/*.image`
do
    # curl -d@${f} 127.0.0.1:13000/files/$(basename ${f})
    curl -s 127.0.0.1:13000/bitcask/$(basename ${f}) | wc -c
done
