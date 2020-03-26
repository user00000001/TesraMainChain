#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             
                                                              

echo "run tesramain in pluto bootnode testnet"
make && \
rm -rf ~/.tesramainchain/internal/tesramain && \
#build/bin/tesramain --nodiscover --internal --port 18822 --syncmode full  --rpc --rpcaddr=0.0.0.0 –rpcapi=’eth,tsr,pos,personal,txpool’ --nodiscover --etherbase  "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e"  --unlock "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e" --password ./pw.txt  --mine --minerthreads=1 $@
build/bin/tesramain --nodiscover --internal --port 18822 --rpc --rpcapi='eth,personal,txpool,pos'   --etherbase  "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e"  --unlock "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e" --password ./pw.txt  --mine --minerthreads=1 $@

