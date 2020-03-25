#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             

 
echo "run tesramain in pluto bootnode testnet for MacOS"
rm -rf ~/Library/TesraMainChain/pluto/tesramain
#make && build/bin/tesramain --pluto --rpc --nodiscover --syncmode "full"
make && build/bin/tesramain --plutodev --nodiscover --etherbase  "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e"  --unlock "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e" --password ~/pw.txt  --mine --minerthreads=1 $@
