#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


nodesNum=$1

if [ $nodesNum -lt 0 ]; then
	return
fi        
                                                            
cd ..

SRCDIR="$(pwd)"

tesramainFile="$SRCDIR/build/bin/tesramain"
if [ ! -x $tesramainFile ]; then
	make all
fi

containerPrefix="tesramainchainContainer"
nodeDirPrefix="$SRCDIR/networkScript/node"
ethbase="0x8b179c2b542f47bb2fb2dc40a3cf648aaae1df16"

allEndNodes=""
port1=8746
port2=3803

for ((i=0;i<$nodesNum;i++));
do
	containerName=$containerPrefix$i
	nodeName="node$i"
	
	let port1++  
	let port2++  

	echo $port1  $port2
	
	cd $SRCDIR/build/bin/
	
	mkdir $SRCDIR/networkScript/$nodeName/
	
	bootnode -genkey $SRCDIR/networkScript/$nodeName/nodekey	
	pubHash=`bootnode -nodekey $SRCDIR/networkScript/$nodeName/nodekey -writeaddress`
	
	if [ ! -x $nodeDirPrefix$i ]; then
		mkdir $nodeDirPrefix$i
	fi
	
	echo "wl" | sudo -S rm $nodeDirPrefix$i/data-loadScript/tesramain
	echo "wl" | sudo -S chmod 777 $nodeDirPrefix$i -R
	
	docker stop $containerName
	docker rm   $containerName
	docker inspect $containerName > /dev/null 2>&1
	if [ $? -eq 1 ]; then
		docker run --restart always --name $containerName -itd -v $SRCDIR:/TesraMainChain/src -p $port1:8545 -p $port2:17717 -p $port2:17717/udp registry.cn-hangzhou.aliyuncs.com/dctech/TesraMainChaindev /bin/sh
	fi
	
	ip=$(docker exec $containerName ifconfig | grep "inet addr" | grep -v 127.0.0.1 | awk '{print $2}' | awk -F ':' '{print $2}')
	endnodeurl="enode://$pubHash@$ip:17717"
	
	if [ $i -eq 0 ]; then
		allEndNodes="$endnodeurl"
	else
		allEndNodes="$allEndNodes,$endnodeurl"
	fi
	
	echo $allEndNodes
 
	cd $SRCDIR

	docker exec -it $containerName /TesraMainChain/src/build/bin/tesramain --datadir "/TesraMainChain/src/networkScript/$nodeName/data-loadScript" init /TesraMainChain/src/genesis_example/genesis.json

	if [ $i -eq 0 ]; then
		echo " start $i"
		docker exec -d $containerName /TesraMainChain/src/build/bin/tesramain --datadir "/TesraMainChain/src/networkScript/$nodeName/data-loadScript" --networkid 314590 --ipcdisable --gasprice 20000 --mine --minerthreads 1 --rpc --rpcaddr 0.0.0.0 --rpcapi "eth,personal,net,admin,tsr" --etherbase $ethbase --nodekey "/TesraMainChain/src/networkScript/$nodeName/nodekey"
	else
		echo " start $i"
		
		docker exec -d $containerName /TesraMainChain/src/build/bin/tesramain --datadir "/TesraMainChain/src/networkScript/$nodeName/data-loadScript" --networkid 314590 --ipcdisable --gasprice 20000 --mine --minerthreads 1 --rpc --rpcaddr 0.0.0.0 --rpcapi "eth,personal,net,admin,tsr" --etherbase $ethbase --nodekey "/TesraMainChain/src/networkScript/$nodeName/nodekey" --bootnodes $allEndNodes
		
	fi

done

