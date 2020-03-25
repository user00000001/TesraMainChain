#!/bin/sh
# set up the logrotate environment to backup wan-chain log data


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


#set logrotate at the miner server in batch

serverUser="ubuntu"
serverPwd=""
serverIps=(
    #"127.0.0.1" #the ip server should be included here
    )
serverKey="/Users/aaron/dc/TesraMainChain_key/TesraMainChain_b.pem.pub"
script="set_logrotate_env.sh"

echo "The log rotate script will be run in batch in servers!\n"

for serverIp in "${serverIps[@]}"
do
    echo "The server " $serverIp " will run the script"

    command=`echo ssh -i $serverKey $serverUser@$serverIp -C \"/bin/bash\"`
    echo $command
    $command < $script

    if [ $? -ne 0 ];then
        echo "The script run with fail"
    else
        echo "The script run successfully"
    fi

    echo ""
done
