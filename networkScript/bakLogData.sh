#!/bin/sh


#  _____                   ____                                   _   
# |_   _|__  ___ _ __ __ _/ ___| _   _ _ __   ___ _ __ _ __   ___| |_ 
#   | |/ _ \/ __| '__/ _` \___ \| | | | '_ \ / _ \ '__| '_ \ / _ \ __|
#   | |  __/\__ \ | | (_| |___) | |_| | |_) |  __/ |  | | | |  __/ |_ 
#   |_|\___||___/_|  \__,_|____/ \__,_| .__/ \___|_|  |_| |_|\___|\__|
#                                     |_|                             


if [ $# != 2 ] ; then
    echo "input 1: $1 log dir"
    echo "input 2: $2 data dir"
    exit 1;
fi

logDir=$1
if [ ! $logDir ] ; then
    echo "log dir is not input"
fi

dataDir=$2
if [ ! $dataDir ] ; then
    echo "data dir is not input"
fi

SRCDIR="$(pwd)"

bakDate=`date +%Y%m%d`
echo $bakDate

ipAddress=$(ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:")
echo $ipAddress

ipStr=`echo $ipAddress | cut -c1-16`

echo $ipStr

bakLogDir="$SRCDIR/backup/log"
bakDataDir="$SRCDIR/backup/data/$ipStr-$bakDate"

echo $bakDataDir
echo $bakDataDir

mkdir -p $bakLogDir
mkdir -p $bakDataDir

cp $logDir"/running.log" $bakLogDir"/"$ipStr"-"$bakDate".log"

cp -r ~/.tesramainchain/* $bakDataDir
