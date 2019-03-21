pragma solidity >=0.4.24 <0.6.0;

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

contract JustitiaRight {
    function inWhiteList(address _account) public view returns(bool);

    function lockCount(address _account, uint _count) public;
    function unlockCount(address _account, uint _count) public;
    function residePledge(address _owner) public view returns(uint balance);

    function balanceOf(address _owner) public view returns (uint256 balance);
    function transfer(address _to, uint256 _value) public returns (bool success);
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool success);
    function approve(address _spender, uint256 _value) public returns (bool success);
    function allowance(address _owner, address _spender) public view returns (uint256 remaining);
}

contract CandidateManage {

    using SafeMath for uint;
    uint public totalPledge;
    uint private gloabalIndex = 0;
    uint constant MINIMUM_PLEDGE_TOKEN = 100;

    JustitiaRight public justitia;

    struct Candidate{
        address account;
        uint pledge;   // total support pledge
        string memo;
        uint ranking;
        bool isValid;
        uint id;
        string url;
    }
    address [] public CandidateList;
    mapping(address => Candidate) public candidateLookup;
    // mapping(support, totalPledge)
    mapping(address => uint256) public balanceOfPledge;

    // event define
    event ApplyToCandidateEvent(address, bool, string);

    // criterias that be a candidate
    function candidateCriteria(address candidate, uint256 pledge) private view returns(bool){
        uint256 balance = justitia.residePledge(candidate);
        if(MINIMUM_PLEDGE_TOKEN <= balance && balance >= pledge){
            return true;
        }
        return false;
    }

    // role classify
    // normal: account with PR, which has right to vote
    // participate: normal account which has participate in vote and has peldge currently
    // candidate: account pledge PR to be a candidate
    function isNormal(address _account) public view returns(bool){
        if(!isParticipate(_account) && !isCandidate(_account)){
            return true;
        }
        return false;
    }

    function isParticipate(address _account) public view returns(bool){
        if (0 != balanceOfPledge[_account]){
            return true;
        }
        return false;
    }

    function isCandidate(address account) public view returns(bool){
        return candidateLookup[account].isValid;
    }

    // get account balance statistic
    function balanceStatistic(address _owner) public view returns (uint256 balance, uint256 pledge){
        require(address(0) != _owner);

        uint256 total;
        uint256 reside;

        total = justitia.balanceOf(_owner);
        reside = justitia.residePledge(_owner);

        require(total.sub(reside) == balanceOfPledge[_owner]);

        return (total, balanceOfPledge[_owner]);
    }

    // get candidate information
    function candidateState(address candidate) public view returns(uint256, uint256, string){
        require(isCandidate(candidate));
        uint index;
        for(index = 0; index < CandidateList.length; index++){
            if(CandidateList[index] == candidate){
                break;
            }
        }
        return (index, candidateLookup[candidate].pledge, candidateLookup[candidate].memo);
    }

    // find index to insert the account by specified candidate in CandidateList
    function findIndexOfCandidate(uint pledge) private view returns(uint){
        uint index;
        for(index = 0; index < CandidateList.length; index++){
            if(candidateLookup[CandidateList[index]].pledge <= pledge){
                break;
            }
        }
        return index;
    }

    // add applicant to candidate list
    function addToCandidateListDescending(address applicant, uint pledge) private returns(uint){
        uint index;
        index = findIndexOfCandidate(pledge);
        CandidateList.push(applicant);
        for(uint i = CandidateList.length - 1; i > index; i--){
            CandidateList[i] = CandidateList[i - 1];
            candidateLookup[CandidateList[i]].ranking = i;
        }
        CandidateList[index] = applicant;
        candidateLookup[CandidateList[index]].ranking = index;
        return index;
    }

    // candidate list adjustment
    function adjustCandidateList(address candidate, uint pledge) public returns(uint){
        if(!isCandidate(candidate)){
            return addToCandidateListDescending(candidate, pledge);
        }

        uint currentIndex;
        uint rightIndex;
        for(currentIndex = 0; currentIndex < CandidateList.length; currentIndex++){
            if(CandidateList[currentIndex] == candidate){
                break;
            }
            if(candidateLookup[CandidateList[rightIndex]].pledge >= candidateLookup[candidate].pledge){
                rightIndex++;
            }

        }
        // adding
        if(rightIndex < currentIndex){
            for(uint i = currentIndex; i > rightIndex; i--){
                CandidateList[i] = CandidateList[i - 1];
                candidateLookup[CandidateList[i]].ranking = i;
            }
        } else {
            for(uint j = currentIndex; j < rightIndex; j++){
                CandidateList[j] = CandidateList[j + 1];
                candidateLookup[CandidateList[j]].ranking = j;
            }
        }

        CandidateList[rightIndex] = candidate;
        candidateLookup[CandidateList[rightIndex]].ranking = rightIndex;
        return rightIndex;
    }

    // apply to candidate
    function ApplyToCandidate(address applicant, uint pledge, string url, string memo) public returns(bool, string){
        require(!isCandidate(applicant));

        string memory errors;
        if(!candidateCriteria(applicant, pledge)){
            errors = "errors: some criterias not met.";
            emit ApplyToCandidateEvent(applicant, false, errors);
            return (false, errors);
        }

        totalPledge = totalPledge.add(pledge);
        justitia.lockCount(applicant, pledge);
        balanceOfPledge[applicant] = balanceOfPledge[applicant].add(pledge);
        adjustCandidateList(applicant, pledge);
        candidateLookup[applicant].memo = memo;
        candidateLookup[applicant].isValid = true;
        candidateLookup[applicant].id = gloabalIndex;
        candidateLookup[applicant].url = url;
        candidateLookup[applicant].pledge = candidateLookup[applicant].pledge.add(pledge);
        candidateLookup[applicant].account = applicant;

        gloabalIndex = gloabalIndex.add(1);
        emit ApplyToCandidateEvent(applicant, true, errors);
        return (true, errors);
    }

    // get candidates
    function Candidates() public view returns(address[]){
        return CandidateList;
    }

    // get participate by ranking
    function GetCandidateByRanking(uint ranking) public view returns(address, uint, string){
        Candidate memory candidate;
        candidate = candidateLookup[CandidateList[ranking]];
        require(candidate.ranking == ranking);
        return (candidate.account, candidate.id, candidate.url);
    }
}

contract BlackListElection {
    using SafeMath for uint;
    struct BlackListItem{
        string reason;
        address []approveAccounts;
        address []rejectAccounts;
        mapping(address => bool) approveRecords;
        mapping(address => bool) rejectRecoeds;
        bool isValid;
        uint index;
    }
    mapping(address => BlackListItem) public blackListItemLookup;
    address [] public blackListProcessing;

    function isRegiste(address _account) public view returns(bool){
        return blackListItemLookup[_account].isValid;
    }

    function toBlackList(address _account, string reason) public returns(uint){
        if(!isRegiste(_account)){
            blackListItemLookup[_account].reason = reason;
            blackListItemLookup[_account].approveAccounts.push(_account);
            blackListItemLookup[_account].approveRecords[msg.sender] = true;
            blackListItemLookup[_account].isValid = true;
            blackListItemLookup[_account].index = blackListProcessing.push(_account).sub(1);
        } else {
            if(!blackListItemLookup[_account].approveRecords[msg.sender]){
                blackListItemLookup[_account].approveAccounts.push(_account);
                blackListItemLookup[_account].approveRecords[msg.sender] = true;
            }
        }
        return blackListItemLookup[_account].approveAccounts.length;
    }

    function blackListToProcess() public view returns(address[]){
        return blackListProcessing;
    }

    function removeBlackList(address _account) public {
        require(isRegiste(_account));
        delete blackListProcessing[blackListItemLookup[_account].index];
        delete blackListItemLookup[_account];
    }
}


contract BlackListManage is BlackListElection{

    uint public thresHoldToAddBlackList;
    uint public thresHoldToRrmoveBlackList;

    struct BlackList {
        uint date;
        bool isValid;
    }
    mapping(address => BlackList) public blackListLookup;
    address [] public blackList;

    event AddToBlackListEvent(address, string);

    function isInBlackList(address _account) public view returns(bool){
        return blackListLookup[_account].isValid;
    }

    function voteForBlacklist(address _account, string comment) public {
        uint supporters = toBlackList(_account, comment);
        if (supporters >=  thresHoldToAddBlackList){
            blackListLookup[_account].date = now;
            blackListLookup[_account].isValid = true;
            blackList.push(_account);
            removeBlackList(_account);
            emit AddToBlackListEvent(_account, blackListItemLookup[_account].reason);
        }
    }

    function getBlackList() public view returns(address[]){
        return blackList;
    }

}

/*
*  系统合约调用
*  管理选举情况，包括：选举，取消选举，选举情况统计等
*/
contract ElectionManage is CandidateManage, BlackListManage {

    using SafeMath for uint;
    uint public totalNodes;
    uint constant ENTRY_HRESHOLD = 100;
    bool private mainNetSwitch;
    uint constant MAINNET_ONLINE_THRESHOLD = 1000;

    event MainNetOnlineEvent(uint, uint);
    event IssueVoteEvent(address, address, uint);
    event AdjustmentVoteEvent(address, address, uint);

    struct Election{
        bool isValid;
        address[] participates;
        mapping(address => uint) election;
    }
    // record candidate election
    mapping(address => Election) private candidateElection;

    // constructor
    constructor () public {
        justitia = JustitiaRight(0xbd770416a3345f91e4b34576cb804a576fa48eb1);
        totalNodes = 4;
        thresHoldToAddBlackList = totalNodes.div(3);
        thresHoldToRrmoveBlackList = thresHoldToAddBlackList.mul(2) + 1;
        ApplyToCandidate(0x333c3310824b7c685133f2bedb2ca4b8b4df633d, 200, "127.0.0.3:47768", "33");
        ApplyToCandidate(0x343c3310824b7c685133f2bedb2ca4b8b4df633d, 199, "127.0.0.4:47768", "34");
        ApplyToCandidate(0x353c3310824b7c685133f2bedb2ca4b8b4df633d, 202, "127.0.0.5:47768", "35");
        ApplyToCandidate(0x363c3310824b7c685133f2bedb2ca4b8b4df633d, 201, "127.0.0.6:47768", "36");
    }

    function setNodeNum(uint nodeNum) public {
        require(isCandidate(msg.sender));
        totalNodes = nodeNum;
    }

    // try to online main network
    // we assume that once mainnet onlie, it will onlie forever
    function tryToOnlineMainNet() private {
        // only state changed, emit event
        if(!mainNetSwitch){
            if(totalPledge >= MAINNET_ONLINE_THRESHOLD){
                mainNetSwitch = true;
                emit MainNetOnlineEvent(now, totalPledge);
            }
        }
    }

    // to get the pledge of participate for specified candidate
    function getPledgeFlow(address candidate, address participate) public view returns(uint){
        require(isCandidate(candidate));
        return candidateElection[candidate].election[participate];
    }

    // to get all supporters of the specified candidate
    function getSupportOfCandidate(address candidate) public view returns(address[]){
        require(isCandidate(candidate));
        return candidateElection[candidate].participates;
    }

    // get rank of candidate
    function ranking(address _candidate) public view returns(uint){
        require(isCandidate(_candidate));
        require(address(0) != _candidate);
        uint index;
        uint rank;
        for(index = 0; index < CandidateList.length; index++){
            if(isCandidate(_candidate)){
                rank++;
                if(_candidate == CandidateList[index]){
                    return rank;
                }
            }
        }
    }

    // issue a election for candidate with some pledges
    function issueVote(address candidate, uint pledge) private {
        require(isCandidate(candidate));
        require(!isCandidate(msg.sender));
        require(pledge <= justitia.residePledge(msg.sender));

        if(!candidateElection[candidate].isValid){
            candidateElection[candidate].participates.push(msg.sender);
            candidateElection[candidate].isValid = true;
        }
        candidateLookup[candidate].pledge = candidateLookup[candidate].pledge.add(pledge);
        candidateElection[candidate].election[msg.sender] = candidateElection[candidate].election[msg.sender].add(pledge);
        adjustCandidateList(candidate, candidateLookup[candidate].pledge);
        balanceOfPledge[msg.sender] = balanceOfPledge[msg.sender].add(pledge);
        totalPledge = totalPledge.add(pledge);
        justitia.lockCount(msg.sender, pledge);

        emit IssueVoteEvent(msg.sender, candidate, pledge);
    }

    // to adjustment of voting for specified candidate with pledge
    function adjustmentVote(address candidate, uint pledge) private {
        require(isCandidate(candidate));
        require(pledge <= candidateElection[candidate].election[msg.sender]);

        candidateLookup[candidate].pledge = candidateLookup[candidate].pledge.sub(pledge);
        candidateElection[candidate].election[msg.sender] = candidateElection[candidate].election[msg.sender].sub(pledge);
        adjustCandidateList(candidate, candidateLookup[candidate].pledge);
        balanceOfPledge[msg.sender] = balanceOfPledge[msg.sender].sub(pledge);
        totalPledge = totalPledge.sub(pledge);
        justitia.unlockCount(msg.sender, pledge);

        emit AdjustmentVoteEvent(msg.sender, candidate, pledge);
    }

    function rightToVoteBlackList(address _account) private view returns(bool){
        require(isCandidate(_account));
        if(candidateLookup[_account].ranking < totalNodes){
            return true;
        }
        return false;
    }

    function GetOnlineSymbol() public view returns(bool){
        return mainNetSwitch;
    }

    function VoteAdjustment(address candidate, uint canceledPledge) public {
        require(isCandidate(candidate));
        adjustmentVote(candidate, canceledPledge);
    }

    function Votting(address candidate, uint pledge) public{
        require(isCandidate(candidate));
        issueVote(candidate, pledge);
        tryToOnlineMainNet();
    }

    function SetBlackList(address _account) public{
        require(rightToVoteBlackList(msg.sender));
        if(!isInBlackList(_account)){
            voteForBlacklist(_account, "errors");
        }
    }
}