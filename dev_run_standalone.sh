#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


echo "run geth"
geth --verbosity 6 --datadir /TesraMainChain/data --etherbase '0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e' --networkid 5201314 --mine --minerthreads 1 --nodiscover --rpc --rpcaddr 0.0.0.0
