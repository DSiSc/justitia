pragma solidity ^0.4.25;


contract WhiteList {
    function inWhiteList(address _account) public view returns(bool);
    function contractProposalStatus(uint proposalId, uint contractId) public returns(bool);
    function changeWhiteListProposalStatus(uint proposalId, address newAddress) public returns(bool);
}

contract MetaData {
    WhiteList public whilteList;

    address private whiteListContractAddress;

    struct contractState{
        uint id;
        bool registered;
        string contractName;
        address contractAddress;
    }
    mapping(uint => contractState) public contractMetaData;

    constructor () public {
        whiteListContractAddress = 0x47e9fbef8c83a1714f1951f142132e6e90f5fa5d;
        whilteList = WhiteList(whiteListContractAddress);
        registerContract(1, 0xbd770416a3345f91e4b34576cb804a576fa48eb1);
        registerContract(2, 0x5a443704dd4b594b382c22a083e2bd3090a6fef3);
        registerContract(3, 0x47e9fbef8c83a1714f1951f142132e6e90f5fa5d);
        registerContract(4, 0x8be503bcded90ed42eff31f56199399b2b0154ca);
    }

    event EventContractRegister(uint, address);
    event EventContraceUpdate(uint, uint, address, address);

    function registerContract(uint _contractId, address _contractAddress) private {
        require(!contractMetaData[_contractId].registered);
        contractMetaData[_contractId].id = _contractId;
        contractMetaData[_contractId].registered = true;
        contractMetaData[_contractId].contractAddress = _contractAddress;
    }

    function updateContract(uint _proposalId, uint _contractId, address _contractNewAddress) public {
        require(contractMetaData[_contractId].registered);
        require(whilteList.contractProposalStatus(_proposalId, _contractId));
        if (contractMetaData[_contractId].contractAddress != _contractNewAddress){
            contractMetaData[_contractId].contractAddress = _contractNewAddress;
            emit EventContraceUpdate(_proposalId, _contractId, contractMetaData[_contractId].contractAddress, _contractNewAddress);
        }
    }

    function getContractById(uint _contractId) public view returns(address){
        require(contractMetaData[_contractId].registered);
        return contractMetaData[_contractId].contractAddress;
    }

    function updateWhiteListAddress(uint _proposalId, address _newAddress) public {
        require(whiteListContractAddress != _newAddress);
        require(whilteList.changeWhiteListProposalStatus(_proposalId, _newAddress));
        whiteListContractAddress = _newAddress;
        whilteList = WhiteList(whiteListContractAddress);
    }
}