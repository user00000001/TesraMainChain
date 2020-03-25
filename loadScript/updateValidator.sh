#!/bin/bash
# set -x
DOCKERIMG=tesramainchain/client-go:2.1.4

echo ''
echo ''
echo ''
echo ''
echo '=========================================='
echo '|   Welcome to testnet Validator Update  |'
echo ''
echo 'If you have deployed your validator with deployValidator.sh, you can update with this script'
echo 'Please make sure that only one tesramain docker is running on the current machine.'
echo 'Otherwise, please update the tesramain version manually.'
echo 'tesramain binary URL: https://github.com/TesraSupernet/TesraMainChain/releases'
echo 'tesramain docker image: ' ${DOCKERIMG}
echo ''
echo ''
echo 'Please Enter your validator Name:'
read YOUR_NODE_NAME
echo 'Please Enter your validator Address'
read addrNew
echo 'Please Enter your password of Validator account:'
read -s PASSWD
echo ''
echo ''
echo ''
echo ''
echo ''


NETWORK=--testnet
NETWORKPATH=testnet

DOCKERID=$(sudo docker ps|grep tesramain|awk '{print $1}')

sudo docker pull ${DOCKERIMG}
if [ $? -ne 0 ]; then
    echo "Docker Pull failed. Please verify your Access of docker command."
    echo "You can add yourself into docker group by this command, and re-login:"
    echo "sudo usermod -aG docker ${USER}"
    exit 1
else
    echo "docker pull succeed"
fi

sudo docker stop ${DOCKERID} >/dev/null 2>&1

sudo docker rm ${DOCKERID} >/dev/null 2>&1

sudo docker stop tesramain >/dev/null 2>&1

sudo docker rm tesramain >/dev/null 2>&1

echo ${PASSWD} | sudo tee -a ~/.tesramainchain/pw.txt > /dev/null
if [ $? -ne 0 ]; then
    echo "Write pw.txt failed"
    exit 1
fi

sudo docker run -d --name tesramain -p 17717:17717 -p 17717:17717/udp -v ~/.tesramainchain:/root/.tesramainchain ${DOCKERIMG} /bin/tesramain ${NETWORK} --etherbase ${addrNew} --unlock ${addrNew} --password /root/.tesramainchain/pw.txt --mine --minerthreads=1 --wanstats ${YOUR_NODE_NAME}:admin@testnet.wanstats.io

if [ $? -ne 0 ]; then
    echo "docker run failed"
    exit 1
fi

echo 'Please wait a few seconds...'

sleep 5

sudo rm ~/.tesramainchain/pw.txt

if [ $? -ne 0 ]; then
    echo "rm pw.txt failed"
    exit 1
fi

echo ''
echo ''
echo ''
echo ''

if [ $(ps -ef | grep -c "tesramain") -gt 1 ]; 
then 
    echo "Validator Start Success";
else
    echo "Validator Start Failed";
    echo "Please use command 'sudo docker logs tesramain' to check reason." 
fi

