function incentiveCheck() {
  count = pos.getIncentiveRunTimes()
  console.log("Incentive count:" + count)

  totalPay = pos.getTotalIncentive()
  remain = pos.getTotalRemain()

  incentiveTotal = 0
  for (i = 0; i < count; i++) {
    incentiveTotal += Number(web3.fromTsl(pos.getIncentivePool(i)[0]))
  }

  console.log("incentive total:" + incentiveTotal)
  console.log("Remain:" + web3.fromTsl(remain))
  console.log("Sub:" + (incentiveTotal - web3.fromTsl(remain)))

  console.log("Sum Pay Total:" + web3.fromTsl(totalPay))

  total = 0
  for (i = 0; i < 4; i++) {
    balance = web3.fromTsl(eth.getBalance(eth.accounts[i]))
    if (balance > 4e12) {
      balance -= 4e13
    } else {
      balance -= 10000000
    }
    console.log(eth.accounts[i] + ":" + balance)
    total += balance
  }

  console.log("Account rised total:" + total)

  gasTotal = 0
  for (i = 0; i < count + 1; i++) {
    gasTotal += Number(web3.fromTsl(pos.getIncentivePool(i)[2]))
  }



  console.log("Gas total:" + gasTotal)

  console.log("TotalGas+AccountIncome:" + (gasTotal + total))
  console.log("compare with:" + web3.fromTsl(totalPay))
  console.log("Deviation:" + (Number(web3.fromTsl(totalPay)) - (gasTotal + total)))

}