#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


echo "run tesramain in testnet"
mkdir -p ./data_testnet
./build/bin/tesramain --testnet --txpool.nolocals --txpool.pricelimit 180000000000 --verbosity 4  --datadir ./data_testnet \
     --rpc --rpcaddr 0.0.0.0 --rpcapi "eth,personal,net,admin,wan" --rpccorsdomain '*' $@
