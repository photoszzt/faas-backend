#!/bin/bash

echo Cleaning storage
sudo rm -rf /var/local/$USER/.linux-img/storage/
sudo mkdir /var/local/$USER/.linux-img/storage/
echo Cleaning logs
sudo chmod g+rx /var/run/firecracker-frakti/
sudo chmod -R g+rx /var/run/firecracker-frakti/default
sudo rm -rf /var/run/firecracker-frakti/default/
sudo mkdir /var/run/firecracker-frakti/default/
