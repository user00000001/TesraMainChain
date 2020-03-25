#!/bin/bash

# cd  ../TesraMainChain-bak
# files=`git diff --name-only`


# for f in ${files[@]}; do
#        cp ${f} ../TesraMainChain/${f}
# done

#cd ../TesraMainChain
#make
#cp ./build/bin/tesramain ../pos6/bin/
#cd ../pos6

docker build . -t tesramainchain/client-go:2.1.2
docker push tesramainchain/client-go:2.1.2