pragma solidity >=0.4.25;

contract CrossFundsPool {
    struct crossTxInfo{
        bytes32 txHash;
        uint txState;
        bool isValid;
    }
    address owner = 0x0001;
    uint256 ticket = 1000000;
    mapping(address => uint256) public funds;
    mapping(address => crossTxInfo) public txnsInfo;

    //deploy contract will first call
    constructor() public {
        //deploy address as owner
        owner = msg.sender;
    }

    function crossTx(string url) public payable returns (bytes32) {
        url;
        require(safeAccount(msg.sender));
        require(msg.value >= ticket);

        funds[msg.sender] = msg.value;
        //TODO: call rpc, cross transaction
        //msg.sender.transfer(msg.value / 2);

        //record crossTxInfo
        bytes32 hash;
        uint state = 0;
        txnsInfo[msg.sender] = crossTxInfo({ txHash: hash, txState: state, isValid: true});

        return  hash;
    }

    function queryTx(address user) public payable returns(bytes32, bool) {
        require(isCrossAccount(user));
        //TODO: query tareget chain tx state


        bool crossTxState = false;
        if (txnsInfo[user].txState == 1) {
            //TODO: if value == 3 ?
            user.transfer(funds[user] / 2);

            funds[user] = 0;
            txnsInfo[user].txHash = bytes32(0);
            txnsInfo[user].txState = 0;
            txnsInfo[user].isValid = false;
            crossTxState = true;
        }

        bytes32 hash = txnsInfo[user].txHash;
        return (hash, crossTxState);
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