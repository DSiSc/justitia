pragma solidity >=0.4.25;

contract RpcContractAddr {
    function ForwardFunds(string toAddr, uint64 amounts, string chainFlag) public view returns (string, uint64) ;
    function GetTxState(string txHash, string chainFlag) public view returns (uint64);
}

contract CrossFundsPool {
    struct crossTxInfo{
        string txHash;
        uint txState;
        bool isValid;
    }
    address owner = 0x0001;
    uint256 ticket = 0;
    mapping(address => uint256) public funds;
    mapping(address => crossTxInfo) public txnsInfo;

    //deploy contract will first call
    constructor() public {
        //deploy address as owner
        owner = msg.sender;
    }

    function test(address contractAddr, string url, uint64 amount) public view {
        RpcContractAddr cross = RpcContractAddr(contractAddr);
        cross.ForwardFunds(url, amount, "chainA");
    }

    function crossTx(address contractAddr, string to, string chainFlag) public payable returns (string) {
        //require(safeAccount(msg.sender));
        //require(msg.value >= ticket);
        funds[msg.sender] = msg.value;

        //TODO: call rpc, cross transaction
        RpcContractAddr cross = RpcContractAddr(contractAddr);
        string memory hash;
        uint64 isOk;
        // rpcAddress 0x0000000000000000000000000000000000011101
        (hash, isOk) = cross.ForwardFunds(to, uint64(msg.value / 2), chainFlag);
        if (isOk != 1) {
            throw;
        }

        //record crossTxInfo
        uint status = 0;
        txnsInfo[msg.sender] = crossTxInfo({ txHash: hash, txState: status, isValid: true});

        return hash;
    }

    function queryTx(address contractAddr, address user, string chainFlag) public payable returns(string, bool) {
        //TODO: query tareget chain tx state
        bool crossTxState = false;
        uint64 state = 0;
        string memory txHash = txnsInfo[user].txHash;

        RpcContractAddr cross = RpcContractAddr(contractAddr);
        state = cross.GetTxState(txHash, chainFlag);

        if (state == 1) {
            //TODO: if value == 3 ?
            user.transfer(funds[user] / 2);

            funds[user] = 0;
            txnsInfo[user].txHash = "0x0000000000000000000000000000000000000000";
            txnsInfo[user].txState = 0;
            txnsInfo[user].isValid = false;
            crossTxState = true;
        } else {
            return (txHash, crossTxState);
        }

        return (txHash, crossTxState);
    }

    //judge the account haven't pending crossTx
    function safeAccount(address addr) public view returns (bool) {
        return (funds[addr]==0) && (!txnsInfo[addr].isValid);
    }

    //judge the address is crossTx account
    function isCrossAccount(address addr) public view returns (bool) {
        return (funds[addr]!=0) && (txnsInfo[addr].isValid);
    }

    //fall back function
    function () public payable { }
}