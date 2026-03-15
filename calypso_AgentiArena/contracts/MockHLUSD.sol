// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MockHLUSD
 * @notice Testnet ERC20 faucet token for HeLa USD
 */
contract MockHLUSD is ERC20 {
    constructor() ERC20("HeLa USD", "HLUSD") {}

    /**
     * @notice Public mint — anyone can call (testnet faucet)
     * @param to Recipient address
     * @param amount Amount to mint (in wei)
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
