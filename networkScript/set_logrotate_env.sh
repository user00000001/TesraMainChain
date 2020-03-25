#!/bin/sh
# set up the logrotate environment to backup wan-chain log data

#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             

#add tesramainchainlog logrotateconf 
version="v1.0.1"
if [ ! -n "$1" ];then
    echo "There is no version parameter input"
    tmp=`ls -lt $HOME/TesraMainChain | grep '^d' | awk '{print $9}' | head -1`
    if [ -f "$HOME/TesraMainChain/$tmp/bin/tesramain" ];then
        version=$tmp
        echo "The newest version is $version"
    fi
else
    echo "The input version parameter is $1"
    version=$1
fi
tesramainchainLogPath=$HOME/TesraMainChain/$version/log/running.log
tesramainchainLogRotateConf=/etc/logrotate.d/TesraMainChainlog

sudo touch $tesramainchainLogRotateConf
sudo chmod 777 $tesramainchainLogRotateConf
echo "
$tesramainchainLogPath
{
   su root list
   daily
   dateext
   rotate 31
   compress
   notifempty
   missingok
   copytruncate
}
" > $tesramainchainLogRotateConf
sudo chmod 644 $tesramainchainLogRotateConf

#add daily schedule to crontab
sudo chmod 777 /etc/crontab
sed -n '/cron.daily/p' /etc/crontab | sudo sed -i 's/25 6/59 23/g' /etc/crontab
sudo chmod 644 /etc/crontab

sudo /etc/init.d/cron restart
