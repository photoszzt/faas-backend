#!/bin/bash

raddr="${RESMNGR_ADDR:-128.83.122.76}"

sudo -E ./fcdaemon -resmngraddr=$raddr "$@"
