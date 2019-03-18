pragma solidity ^0.4.25;

contract Test {

    string public greeting;

    constructor() public {
        greeting = "Hello, world.";
    }

    function greet() public view returns (string memory) {
        return greeting;
    }

    function setGreeting(string memory _greeting) public {
        greeting = _greeting;
    }
}