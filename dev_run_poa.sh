#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


echo "run tesramain in pluto testnet"
mkdir -p ./data_pluto
echo "dc" > ./passwd.txt
if [ -d "DOCKER" ]; then
	cp -rf DOCKER/data/keystore ./data_pluto
else
	cp -rf ../data/keystore ./data_pluto
fi
#rm -rf ./data_pluto/tesramain/chaindata
#tesramain --datadir ./data_pluto init ./genesis_example/genesis_poa.json
networkid='--pluto'
./build/bin/tesramain ${networkid} --nat none --verbosity 4 --gasprice '200000' --datadir ./data_pluto  \
	 --unlock "0x2d0e7c0813a51d3bd1d08246af2a8a7a57d8922e" --password ./passwd.txt \
	 --targetgaslimit 900000000  --port 17717  \
   	--rpc --rpcaddr 0.0.0.0 --rpcapi "eth,personal,net,admin" --rpccorsdomain '*' $@
