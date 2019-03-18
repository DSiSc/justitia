pragma solidity >=0.4.24 <0.6.0;

/* 安全操作函数
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


/* 权限管理接口
 * Ownable provides basic authorization control to simplifies the implementation of "user permissions".
 */
contract Ownable {
    address public owner;
    /*
    * The Ownable constructor sets the original `owner` of the contract to the sender account which deploy the account.
    */
    constructor () public {
        owner = msg.sender;
    }
    /*
    * Throws if called by any account other than the owner.
    */
    modifier onlyOwner() {
        if (msg.sender != owner) {
            revert();
        }
        _;
    }
    modifier validAddress {
        assert(0x0 != msg.sender);
        _;
    }
    /*
    * Allows the current owner to transfer control of the contract to a newOwner.
    * @param newOwner The address to transfer ownership to.
    */
    function transferOwnership(address newOwner) onlyOwner validAddress public {
        if (newOwner != address(0)) {
            owner = newOwner;
        }
    }
}

/* 设置全局锁仓接口：无论任何账户，在锁仓情况向都不能进行交易
 * 是否在锁仓状态下开启owner用户和admin用户的权限
 */
contract Lockable is Ownable {
    using SafeMath for uint;
    bool public lockStatus = false;
    mapping (address => bool) accountLockStatus;
    mapping (address => uint) lockStatistic;

    event Lock(address);
    event UnLock(address);
    event AccountLocked(address, bool);

    /* 条件验证：非锁仓状态 : 条件： 全局未锁仓，*/
    modifier unLocked() {
        if (lockStatus){
            revert();
        }
        if (accountLockStatus[msg.sender]){
            revert();
        }
        _;
    }

    /* 条件验证：锁仓状态 */
    modifier inLocked() {
        if (lockStatus){
            if (accountLockStatus[msg.sender]){
                revert();
            }
        }
        _;
    }

    /* 设置锁仓 */
    function lock() onlyOwner unLocked public returns (bool) {
        lockStatus = true;
        emit Lock(msg.sender);
        return true;
    }

    /* 解锁 */
    function unlock() onlyOwner inLocked public returns (bool) {
        lockStatus = false;
        emit UnLock(msg.sender);
        return true;
    }

    /* 查询全局锁状态*/
    function getLockStatus() validAddress public view returns (bool){
        return lockStatus;
    }

    /* 锁定指定账户 */
    function lockAccount(address _account) onlyOwner public returns (bool) {
        require(address(0) != _account);
        accountLockStatus[_account] = true;
        emit AccountLocked(_account, true);
        return true;
    }

    /* 解锁指定账户 */
    function unlockAccount(address _account) onlyOwner public returns (bool) {
        require(address(0) != _account);
        accountLockStatus[_account] = false;
        emit AccountLocked(_account, false);
        return true;
    }

    /* 查询指定账户单独锁定的状态，不返回全局锁 */
    function getAccountLockStatus(address _account) validAddress public view returns (bool){
        require(address(0) != _account);
        return accountLockStatus[_account];
    }

    /* 锁定指定账户的指定数量的JR */
    function lockCount(address _account, uint _count) public {
        require(address(0) != _account);
        lockStatistic[_account] = lockStatistic[_account].add(_count);
    }

    /* 解锁指定账户的指定数量JR */
    function unlockCount(address _account, uint _count) public {
        require(address(0) != _account);
        lockStatistic[_account] = lockStatistic[_account].sub(_count);
    }
}

/*
 * ERC20 interface
 * see https://github.com/ethereum/EIPs/issues/20
 */
contract ERC20 {
    function balanceOf(address _owner) view public returns (uint256 balance);
    function transfer(address _to, uint256 _value) public returns (bool success);
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool success);
    function approve(address _spender, uint256 _value) public returns (bool success);
    function allowance(address _owner, address _spender) public view returns (uint256 remaining);
    event Transfer(address indexed _from, address indexed _to, uint256 _value);
    event Approval(address indexed _owner, address indexed _spender, uint256 _value);
}

/* 支持抵押功能的ERC20模型
 * @title Standard ERC20 token with pledge function
 * @dev Implemantation of the basic standart token.
 * @dev https://github.com/ethereum/EIPs/issues/20
 * @dev Based on code by FirstBlood: https://github.com/Firstbloodio/token/blob/master/smart_contract/FirstBloodToken.sol
 */
contract JRBase is ERC20,Lockable{
    using SafeMath for uint;
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;
    mapping (address => uint256) public balanceOf;
    mapping (address => mapping (address => uint256)) public allowance;

    /* 查询余额 */
    function balanceOf(address _owner) view public returns (uint256 balance){
        require(address(0) != _owner);
        return (balanceOf[_owner]);
    }

    // 查询可抵押余额：返回可用于抵押的额度
    function residePledge(address _owner) public view returns(uint balance){
        return balanceOf[_owner].sub(lockStatistic[_owner]);
    }

    /* 查询授权额度 */
    function allowance(address _owner, address _spender) view public returns (uint256 remaining) {
        require(address(0) != _owner);
        require(address(0) != _spender);
        return allowance[_owner][_spender];
    }

    /* 转账函数*/
    function transfer(address _to, uint256 _value) public returns (bool success) {
        require(residePledge(_to) >= _value);
        balanceOf[msg.sender] = balanceOf[msg.sender].sub(_value);
        balanceOf[_to] = balanceOf[_to].add(_value);
        emit Transfer(msg.sender, _to, _value);
        return true;
    }

    /* 授权转账函数 */
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool success) {
        require(residePledge(_from) >= _value);
        balanceOf[_to] = balanceOf[_to].add(_value);
        balanceOf[_from] = balanceOf[_from].sub(_value);
        allowance[_from][msg.sender] = allowance[_from][msg.sender].sub(_value);
        emit Transfer(_from, _to, _value);
        return true;
    }

    /* 授权接口，调整授权额度 */
    function approve(address _spender, uint256 _value) public returns (bool success) {
        require(address(0) != _spender);
        allowance[msg.sender][_spender] = _value;
        emit Approval(msg.sender, _spender, _value);
        return true;
    }
}

contract JustitiaRight is JRBase {
    uint public ratio;
    event IssueEvent(address _to, uint _balance);

    // 总的发行量
    // constructor (string _name, string _symbol, uint8 _decimals) public {
    constructor () public {
        totalSupply = 0;
        name = "Justitia Repution Token";
        symbol = "JR";
        decimals = 18;
        ratio = 1;
        balanceOf[msg.sender] = totalSupply;
        balanceOf[0x86edb13de37acc08110a3516e09b762245254b24] = 10000;
        balanceOf[0xff0f61bc6044021512ebd584dd26e74fb38a4928] = 10000;
        balanceOf[0x563ce264a98480c0c992431737b2d23b046b71b7] = 10000;
        balanceOf[0xb39916591fb877e55e1fe647372cdbddccc6c3d3] = 10000;
        totalSupply = totalSupply + 40000;
    }

    // 直接发行JR，用于挖矿奖励以及其他奖励
    function issueJR(address _to, uint _balance) public returns(uint){
        balanceOf[_to] = balanceOf[_to].add(_balance);
        totalSupply = totalSupply.add(_balance);
        emit IssueEvent(_to, _balance);
    }

    // 一定数量的JT可兑换的JR数量
    function jtToJR(uint _count) private view returns(uint){
        return _count.mul(ratio);
    }

    // 使用以太购买JR
    function buyJR() payable public {
        require(0 != msg.value);
        issueJR(msg.sender, msg.value);
    }

    // 获取指定账户的以太余额
    function getEth() public view returns (uint){
        return address(this).balance;
    }

    /*
    // 提取以太
    function withdrawEth(uint256 amount) public onlyOwner {
        owner.transfer(amount);
    }

    // 设置代理取走以太
    function withdrawEthByProxy(address admin, uint256 amount) public {
        admin.transfer(amount);
    }
    */

}