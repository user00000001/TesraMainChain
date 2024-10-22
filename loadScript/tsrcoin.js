var tsrBalance = function(addr){
	return web3.fromTsl(web3.eth.getBalance(addr));
}

var tsrUnlock = function(addr){
    return personal.unlockAccount(addr,"dc",99999);	
}

var sendTsrFromUnlock = function (From, To , V){
	return eth.sendTransaction({from:From, to: To, value: web3.toTsl(V)});
}

var wait = function (conditionFunc) {
	var loopLimit = 100;
	var loopTimes = 0;
	while (!conditionFunc()) {
		admin.sleep(1);
		loopTimes++;
		if(loopTimes>=loopLimit){
			throw Error("wait timeout! conditionFunc:" + conditionFunc)
		}
	}
}

var tranValue = 10;

abiDef = [{"constant":false,"type":"function","stateMutability":"nonpayable","inputs":[{"name":"OtaAddr","type":"string"},{"name":"Value","type":"uint256"}],"name":"buyCoinNote","outputs":[{"name":"OtaAddr","type":"string"},{"name":"Value","type":"uint256"}]},{"constant":false,"type":"function","inputs":[{"name":"RingSignedData","type":"string"},{"name":"Value","type":"uint256"}],"name":"refundCoin","outputs":[{"name":"RingSignedData","type":"string"},{"name":"Value","type":"uint256"}]},{"constant":false,"inputs":[],"name":"getCoins","outputs":[{"name":"Value","type":"uint256"}]}];

contractDef = eth.contract(abiDef);
coinContractAddr = "0x0000000000000000000000000000000000000064";
coinContract = contractDef.at(coinContractAddr);

tsrUnlock(eth.accounts[1]);
tsrUnlock(eth.accounts[2]);

for (i = 0; i < 3; i++) {
    var tsrAddr = tsr.getTsrAddress(eth.accounts[2]);
    var otaAddr = tsr.generateOneTimeAddress(tsrAddr);

    txBuyData = coinContract.buyCoinNote.getData(otaAddr, web3.toTsl(tranValue));
    buyCoinTx = eth.sendTransaction({from:eth.accounts[1], to:coinContractAddr, value:web3.toTsl(tranValue), data:txBuyData, gas: 1000000, gasprice:'0x' + (20000000000).toString(16)});
    wait(function(){return eth.getTransaction(buyCoinTx).blockNumber != null;});
}

var acc1OldBalance = parseFloat(tsrBalance(eth.accounts[1]))
var acc2OldBalance = parseFloat(tsrBalance(eth.accounts[2]))


var tsrAddr = tsr.getTsrAddress(eth.accounts[2]);
var otaAddr = tsr.generateOneTimeAddress(tsrAddr);

txBuyData = coinContract.buyCoinNote.getData(otaAddr, web3.toTsl(tranValue));
buyCoinTx = eth.sendTransaction({from:eth.accounts[1], to:coinContractAddr, value:web3.toTsl(tranValue), data:txBuyData, gas: 1000000, gasprice:'0x' + (20000000000).toString(16)});
wait(function(){return eth.getTransaction(buyCoinTx).blockNumber != null;});


var mixTsrAddresses = tsr.getOTAMixSet(otaAddr,2);
var mixSetWith0x = []
for (i = 0; i < mixTsrAddresses.length; i++){
	mixSetWith0x.push(mixTsrAddresses[i])
}

keyPairs = tsr.computeOTAPPKeys(eth.accounts[2], otaAddr).split('+');
privateKey = keyPairs[0];

console.log("Balance of ", eth.accounts[2], " is ", web3.fromTsl(eth.getBalance(eth.accounts[2])));
var ringSignData = personal.genRingSignData(eth.accounts[2], privateKey, mixSetWith0x.join("+"))
var txRefundData = coinContract.refundCoin.getData(ringSignData, web3.toTsl(tranValue))
var refundTx = eth.sendTransaction({from:eth.accounts[2], to:coinContractAddr, value:0, data:txRefundData, gas: 2000000, gasprice:'0x' + (20000000000).toString(16)});
wait(function(){return eth.getTransaction(refundTx).blockNumber != null;});

console.log("New balance of ", eth.accounts[2], " is ", web3.fromTsl(eth.getBalance(eth.accounts[2])));

var acc1NewBalance = parseFloat(tsrBalance(eth.accounts[1]))
var acc2NewBalance = parseFloat(tsrBalance(eth.accounts[2]))
if (acc2NewBalance < acc2OldBalance || acc2NewBalance > (acc2OldBalance + tranValue)) {
	throw Error("acc2OldBalance:" + acc2OldBalance + ", acc2NewBalance:" + acc2NewBalance + ", tranValue:" + tranValue)
}

if (acc1NewBalance > acc1OldBalance - tranValue || acc1NewBalance < acc1OldBalance - tranValue - 1) {
	throw Error("acc1OldBalance:" + acc1OldBalance + ", acc1NewBalance:" + acc1NewBalance + ", tranValue:" + tranValue)
}


