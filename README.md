# Justitia

[![Build Status](https://circleci.com/gh/DSiSc/justitia/tree/master.svg?style=shield)](https://circleci.com/gh/DSiSc/justitia/tree/master)
[![codecov](https://codecov.io/gh/DSiSc/justitia/branch/master/graph/badge.svg)](https://codecov.io/gh/DSiSc/justitia)

Project Justitia builds a global community and a code library of the smart contract,
relies on crowdfunding and crowd-developing on Justitia Chain to reach a broad consensus,
and is committed to making code gradually become an "invisible law" in the virtual world,
forming a trustworthy contract production mechanism. 
This project uses innovative Law Code technology and smart contract engineering theory
to create a crowd-developed contract production DAC community DSiSc(DAC Swarm intelligence
community of Smart contract), designing and implementing a new Justitia public blockchain
to achieve community governance, and inspiring and bringing together global intelligence
consensus to build a smart contract production factory.

Smart contracts will be important basic protocols with legal attributes in the digital
society in the future and will become important foundations for the digital society in
the future. We believe that smart contracts will be an important technological factor in
the process of the blockchain revolution. However, smart contracts that are executed
automatically as treaties and rules face many unprecedented challenges. 
Issues of letting smart contracts carry legal rules, store certificates, and automatically
judge program execution will become a great challenge in the process of universal
application of smart contracts.

## Getting started

Here's how to set up `justitia` for local development.

1. Fork the `justitia` repo on GitHub.
2. Clone your fork locally, and fetch all dependencies:

        $ git clone git@github.com:your_name_here/justitia.git
        $ cd justitia
        $ make fetch-deps

3. Create a branch for local development:

        $ git checkout -b name-of-your-bugfix-or-feature

   Now you can make your changes locally.

4. When you're done making changes, check that your changes pass the tests:

        $ make test
        
5. Or directly run justitia with the following command:

        $ go run main.go

6. Commit your changes and push your branch to GitHub, We use [Angular Commit Guidelines](https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#-git-commit-guidelines), Thanks for Angular good job.::

        $ git add .
        $ git commit -m "Your detailed description of your changes."
        $ git push origin name-of-your-bugfix-or-feature

7. Submit a pull request through the GitHub website.

## Sub-projects

- [apigateway](https://github.com/DSiSc/apigateway): A light-weight golang API Gateway implement.
- [wallet](https://github.com/DSiSc/wallet): A high-security wallet implemention.
- [crypto-suite](https://github.com/DSiSc/crypto-suite): Crypto Suite.
- [craft](https://github.com/DSiSc/craft): Define common types and structures which used frequently.
- [blockstore](https://github.com/DSiSc/blockstore): An implemention of ledger which support customization by config file.
- [blockchain](https://github.com/DSiSc/blockchain): Middleware of blockchain storage layer accessing.
- [evm-NG](https://github.com/DSiSc/evm-NG): Next Generation Contract VM from EVM.
- [statedb-NG](https://github.com/DSiSc/statedb-NG): A Next Generation StateDB Implemention.
- [gossipswitch](https://github.com/DSiSc/gossipswitch): A Gossip switch implementation.
- [validator](https://github.com/DSiSc/validator): A high-speed validator verify transaction and block.
- [txpool](https://github.com/DSiSc/txpool): A high-performance blockchain transaction pool.
- [producer](https://github.com/DSiSc/producer): Implement of producer which is responsible for producing block.
- [galaxy](https://github.com/DSiSc/galaxy): Advanced distributed consensus framework support pluggable algorithms.


## Releases

- [v0.1 - Sep 11, 2018](https://github.com/DSiSc/justitia/releases/tag/v0.1)

## Licensing

Justitia is licensed under the Apache License, Version 2.0. See
[LICENSE](https://github.com/moby/moby/blob/master/LICENSE) for the full
license text.

