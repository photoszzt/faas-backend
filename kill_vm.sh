#!/bin/bash
sudo pkill -KILL fcdaemon
sudo pkill -KILL firecracker
sudo rm -rf /var/local/$USER/.capstan/storage/*
sudo rm -rf /var/local/$USER/.linux-img/storage/*
