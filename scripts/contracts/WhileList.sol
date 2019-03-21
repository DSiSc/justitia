pragma solidity ^0.4.25;

/*
* 安全操作函数
*  SafeMath to avoid data overwrite
*/
library SafeMath {
    function mul(uint a, uint b) internal pure returns (uint) {
        uint c = a * b;
        require(a == 0 || c / a == b, "overwrite error");
        return c;
    }

    function div(uint a, uint b) internal pure returns (uint) {
        require(b > 0, "overwrite error");
        uint c = a / b;
        require(a == b * c + a % b, "overwrite error");
        return c;
    }

    function sub(uint a, uint b) internal pure returns (uint) {
        require(b <= a, "overwrite error");
        return a - b;
    }

    function add(uint a, uint b) internal pure returns (uint) {
        uint c = a + b;
        require(c>=a && c>=b, "overwrite error");
        return c;
    }
}

contract WhiteList {

    using SafeMath for uint;

    uint DefaultOpcode = 0;
    uint AddOpcode = 1;
    uint DeleteOpcode = 2;
    struct WhiteListProposal {
        uint proposalId;
        address participate;
        address issueAddress;
        uint opcode;
        uint votes;
        bool isExist;
        bool over;
        uint currentVotes;
        mapping(address => bool) voteState;
    }
    uint public totalParticiates = 0;
    mapping(address => uint) participateOpcodeState;
    address [] public whiteList;
    mapping(address => bool) private whiteListState;
    mapping(uint => WhiteListProposal) private whiteListProposalState;
    event EventAddToWhiteList(uint, address);
    event EventRemoveFromWhiteList(uint, address);

    constructor() public {
        address defaultAccount = 0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b;
        //for(uint index = 0; index < defaultAccount.length; index++){
        //    whiteList.push(defaultAccount[index]);
        //    whiteListState[defaultAccount[index]] = true;
        //    totalParticiates = totalParticiates.add(1);
        //}
        whiteList.push(defaultAccount);
        whiteListState[defaultAccount] = true;
        totalParticiates = totalParticiates.add(1);
    }

    function inWhiteList(address _account) public view returns(bool) {
        return whiteListState[_account];
    }

    function issueWhileListProposal(uint proposalId, address _account, uint _opcode) public {
        require(inWhiteList(msg.sender));
        require(!inWhiteList(_account));
        require(!whiteListProposalState[proposalId].isExist);
        // current state
        require(DefaultOpcode == participateOpcodeState[_account]);

        participateOpcodeState[_account] = _opcode;
        whiteListProposalState[proposalId].proposalId = proposalId;
        whiteListProposalState[proposalId].participate = _account;
        whiteListProposalState[proposalId].opcode = _opcode;
        whiteListProposalState[proposalId].issueAddress = msg.sender;
        whiteListProposalState[proposalId].voteState[msg.sender] = true;
        whiteListProposalState[proposalId].currentVotes = whiteListProposalState[proposalId].currentVotes.add(1);
    }

    function voteForWhiteListProposal(uint proposalId) public {
        require(inWhiteList(msg.sender));
        require(whiteListProposalState[proposalId].isExist);
        require(!whiteListProposalState[proposalId].voteState[msg.sender]);
        require(!whiteListProposalState[proposalId].over);
        if (AddOpcode == whiteListProposalState[proposalId].opcode){
            addToWhiteList(proposalId, msg.sender, whiteListProposalState[proposalId].participate);
        } else {
            removeFromWhiteList(proposalId, msg.sender, whiteListProposalState[proposalId].participate);
        }
    }

    function findRecordIndex(address _address) private view returns(uint){
        for(uint index = 0; index < whiteList.length; index++){
            if(whiteList[index] == _address){
                return index;
            }
        }
    }

    function conditionsForWhiteList(uint proposalId) private view returns(bool) {
        uint threshold = totalParticiates.div(3).mul(2) + 1;
        if (whiteListProposalState[proposalId].currentVotes >= threshold) {
            return true;
        }
        return false;
    }

    function addToWhiteList(uint proposalId, address _sender, address participate) private {
        require(!inWhiteList(participate));
        whiteListProposalState[proposalId].voteState[_sender] = true;
        whiteListProposalState[proposalId].currentVotes = whiteListProposalState[proposalId].currentVotes.add(1);
        if (conditionsForWhiteList(proposalId)) {
            whiteList.push(whiteListProposalState[proposalId].participate);
            whiteListState[whiteListProposalState[proposalId].participate] = true;
            whiteListProposalState[proposalId].over = true;
            participateOpcodeState[whiteListProposalState[proposalId].participate] = DefaultOpcode;
            totalParticiates = totalParticiates.add(1);
            emit EventAddToWhiteList(proposalId, whiteListProposalState[proposalId].participate);
        }
    }

    function removeFromWhiteList(uint proposalId, address _sender, address participate) private {
        require(inWhiteList(participate));
        whiteListProposalState[proposalId].voteState[_sender] = true;
        whiteListProposalState[proposalId].currentVotes = whiteListProposalState[proposalId].currentVotes.add(1);
        if (conditionsForWhiteList(proposalId)) {
            uint index = findRecordIndex(participate);
            delete whiteList[index];
            whiteListState[participate] = true;
            whiteListProposalState[proposalId].over = true;
            totalParticiates = totalParticiates.sub(1);
            participateOpcodeState[participate] = DefaultOpcode;
            emit EventRemoveFromWhiteList(proposalId,participate);
        }
    }

    struct ContractProposal {
        uint proposalId;
        uint contractId;
        uint currentVotes;
        bool isExist;
        bool over;
        address issueAddress;
        mapping(address => bool) voteState;
    }
    mapping(uint => ContractProposal) public contractProposalState;
    mapping(uint => bool) public contractProposalCalledState;
    event EventIssueContractProposal(address, uint, uint);
    event EventVoteContractProposal(address, uint);

    function issueContractProposal(uint proposalId, uint contractId) public {
        require(!contractProposalState[proposalId].isExist);
        contractProposalState[proposalId].proposalId = proposalId;
        contractProposalState[proposalId].contractId = contractId;
        contractProposalState[proposalId].issueAddress = msg.sender;
        contractProposalState[proposalId].isExist = true;
        if (inWhiteList(msg.sender)) {
            contractProposalState[proposalId].voteState[msg.sender] = true;
            contractProposalState[proposalId].currentVotes = contractProposalState[proposalId].currentVotes.add(1);
        }
        emit EventIssueContractProposal(msg.sender, proposalId, contractId);
    }

    function conditionsForContract(uint proposalId) private view returns(bool){
        if (contractProposalState[proposalId].isExist){
            uint thresHold = totalParticiates.div(3).mul(2) + 1;
            if (contractProposalState[proposalId].currentVotes >= thresHold) {
                return true;
            }
        }
        return false;
    }


    function voteForContractProposal(uint proposalId) public {
        require(inWhiteList(msg.sender));
        require(contractProposalState[proposalId].isExist);
        require(!contractProposalState[proposalId].voteState[msg.sender]);
        require(!contractProposalState[proposalId].over);

        contractProposalState[proposalId].voteState[msg.sender] = true;
        contractProposalState[proposalId].currentVotes = contractProposalState[proposalId].currentVotes.add(1);
        if (conditionsForContract(proposalId)){
            contractProposalState[proposalId].over = true;
        }
        emit EventVoteContractProposal(msg.sender, proposalId);
    }

    function contractProposalStatus(uint proposalId, uint contractId) public returns(bool){
        // only can be called once
        if (!contractProposalCalledState[proposalId]){
            if (contractProposalState[proposalId].contractId == contractId && contractProposalState[proposalId].over) {
                contractProposalCalledState[proposalId] = true;
                return true;
            }
        }
        return false;
    }

    function getContratProposal(uint proposalId) public view returns(uint, address, uint){
        require(contractProposalState[proposalId].isExist);
        return (contractProposalState[proposalId].contractId, contractProposalState[proposalId].issueAddress, contractProposalState[proposalId].currentVotes);
    }

    struct ChangeWhiteListProposal {
        uint proposalId;
        address originAddress;
        address newAddress;
        bool isExist;
        bool over;
        uint currentVotes;
        address issueAddress;
        mapping(address => bool) voteState;
    }
    mapping(uint => ChangeWhiteListProposal) public changeWhiteListProposalState;
    mapping(uint => bool) public changeWhiteListProposalCalledState;

    event EventIssueChangeWhiteListProposal(address, uint, address, address);
    event EventVoteForChangeWhiteListProposal(address, uint);

    function issueChangeWhiteListProposal(uint proposalId, address originAddress, address newAddress) public {
        require(!changeWhiteListProposalState[proposalId].isExist);
        changeWhiteListProposalState[proposalId].proposalId = proposalId;
        changeWhiteListProposalState[proposalId].originAddress = originAddress;
        changeWhiteListProposalState[proposalId].newAddress = newAddress;
        changeWhiteListProposalState[proposalId].isExist = true;
        if (inWhiteList(msg.sender)) {
            changeWhiteListProposalState[proposalId].voteState[msg.sender] = true;
            changeWhiteListProposalState[proposalId].currentVotes = changeWhiteListProposalState[proposalId].currentVotes.add(1);
        }
        emit EventIssueChangeWhiteListProposal(msg.sender, proposalId, originAddress, newAddress);
    }

    function conditionsForChangeWhiteListProposal(uint proposalId) private view returns(bool){
        if (changeWhiteListProposalState[proposalId].isExist){
            uint thresHold = totalParticiates.div(3).mul(2) + 1;
            if (changeWhiteListProposalState[proposalId].currentVotes >= thresHold) {
                return true;
            }
        }
        return false;
    }

    function voteForChangeWhiteListProposal(uint proposalId) public {
        require(inWhiteList(msg.sender));
        require(changeWhiteListProposalState[proposalId].isExist);
        require(!changeWhiteListProposalState[proposalId].voteState[msg.sender]);
        require(!changeWhiteListProposalState[proposalId].over);

        changeWhiteListProposalState[proposalId].voteState[msg.sender] = true;
        changeWhiteListProposalState[proposalId].currentVotes = changeWhiteListProposalState[proposalId].currentVotes.add(1);
        if (conditionsForChangeWhiteListProposal(proposalId)){
            changeWhiteListProposalState[proposalId].over = true;
        }
        emit EventVoteForChangeWhiteListProposal(msg.sender, proposalId);
    }

    function changeWhiteListProposalStatus(uint proposalId, address newAddress) public returns(bool){
        // only can be called once
        if (!changeWhiteListProposalCalledState[proposalId]){
            if (changeWhiteListProposalState[proposalId].newAddress == newAddress && changeWhiteListProposalState[proposalId].over) {
                changeWhiteListProposalCalledState[proposalId] = true;
                return true;
            }
        }
        return false;
    }

}