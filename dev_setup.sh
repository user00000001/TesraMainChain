#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


SRCDIR="$(pwd)"
docker inspect tesramainchainContainer > /dev/null 2>&1
if [ $? -eq 1 ]; then
	docker run --restart always --name tesramainchainContainer -itd -v $SRCDIR:/TesraMainChain/src -p 8545:8545 -p 17717:17717 -p 17717:17717/udp  registry.cn-hangzhou.aliyuncs.com/dctech/TesraMainChaindev /bin/sh
fi
docker exec -it tesramainchainContainer /bin/sh

