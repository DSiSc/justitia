pragma solidity >=0.4.25;

contract RpcContractAddr {
    function ForwardFunds(string toAddr, uint64 amounts, string payload, string chainFlag) public view returns (string, string, uint64) ;
    function GetTxState(string txHash, uint64 amount, string tmp, string chainFlag) public view returns (uint64);
    function ReceiveFunds(address user, uint64 amount, string payload, uint64 chainId) public view returns (uint64);
}

contract CrossFundsPool {
    struct crossTxInfo{
        address toAddr;
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

    function crossTx(address to, string payload, string chainFlag) public payable returns (string) {
        //require(safeAccount(msg.sender));
        //require(msg.value >= ticket);
        address contractAddr = 0x0000000000000000000000000000000000011101;

        //msg.sender is tx's from ???
        funds[msg.sender] = msg.value;

        //TODO: call rpc, cross transaction
        RpcContractAddr cross = RpcContractAddr(contractAddr);
        string memory localHash;
        string memory targetHash;
        uint64 isOk;
        string memory toAddr;

        toAddr = toAsciiString(to);
        (localHash, targetHash, isOk) = cross.ForwardFunds(toAddr, uint64(msg.value), payload, chainFlag);
        if (isOk != 1) {
            throw;
        }

        //record crossTxInfo
        uint status = 0;
        txnsInfo[msg.sender] = crossTxInfo({toAddr: to, txHash: targetHash, txState: status, isValid: true});

        return targetHash;
    }

    function queryTx(address user, string chainFlag) public payable returns(string, bool) {
        bool crossTxState = false;
        uint64 state = 0;
        string memory txHash = txnsInfo[user].txHash;
        address contractAddr = 0x0000000000000000000000000000000000011101;
        string memory tmp;
        tmp = toAsciiString(user);

        RpcContractAddr cross = RpcContractAddr(contractAddr);
        state = cross.GetTxState(txHash, uint64(funds[user]), tmp, chainFlag);

        if (state == 1) {
            //TODO: if value == 3 ?
            require(funds[user] > 0);
            user.transfer(funds[user]);

            funds[user] = 0;
            txnsInfo[user].toAddr = 0x0000000000000000000000000000000000000000;
            txnsInfo[user].txHash = "0x0000000000000000000000000000000000000000";
            txnsInfo[user].txState = 0;
            txnsInfo[user].isValid = false;
            crossTxState = true;
        } else {
            return (txHash, crossTxState);
        }

        return (txHash, crossTxState);
    }

    function receiveFunds(address user, string payload, uint64 amount, uint64 chainId) public payable returns(bool){
        uint64 state = 0;
        address contractAddr = 0x0000000000000000000000000000000000011101;
        string memory toString;

        RpcContractAddr cross = RpcContractAddr(contractAddr);
        state = cross.ReceiveFunds(user, amount, payload, chainId);
        if (state == 1) {
            user.transfer(amount);
            return true;

        } else {
            return false;
        }
    }

    //judge the account haven't pending crossTx
    function safeAccount(address addr) public view returns (bool) {
        return (funds[addr]==0) && (!txnsInfo[addr].isValid);
    }

    //judge the address is crossTx account
    function isCrossAccount(address addr) public view returns (bool) {
        return (funds[addr]!=0) && (txnsInfo[addr].isValid);
    }

    function toAsciiString(address x) returns (string) {
        bytes memory s = new bytes(40);
        for (uint i = 0; i < 20; i++) {
            byte b = byte(uint8(uint(x) / (2**(8*(19 - i)))));
            byte hi = byte(uint8(b) / 16);
            byte lo = byte(uint8(b) - 16 * uint8(hi));
            s[2*i] = char(hi);
            s[2*i+1] = char(lo);
        }

        return string(s);
    }

    function char(byte b) returns (byte c) {
        if (b < 10) return byte(uint8(b) + 0x30);
        else return byte(uint8(b) + 0x57);
    }

    //fall back function
    function () public payable { }
}