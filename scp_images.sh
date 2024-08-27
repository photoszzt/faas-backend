#!/bin/bash

if (( $# != 1 )); then
    echo "need ip to copy"
fi

#ssh ubuntu@$1 "sudo rm -rf /var/local/ubuntu/.linux-img/*"
#rsync -rvhe ssh --progress --exclude 'storage/*' /var/local/ubuntu/.linux-img/ ubuntu@$1:/var/local/ubuntu/.linux-img

stages=(parlink decode grayscale encode)
#stages=(grayscale encode) 

for stage in ${stages[@]}; do
    ssh ubuntu@$1 "sudo rm -rf /var/local/ubuntu/.linux-img/${stage}"
    scp -r /var/local/ubuntu/.linux-img/$stage ubuntu@$1:/var/local/ubuntu/.linux-img/
done
