// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";

interface IAgentRegistryForRep {
    function suspendAgent(uint256 agentId) external;
    function agents(uint256 agentId) external view returns (
        uint256 id, string memory name, string memory category,
        string memory framework, uint256 pricePerCall,
        address payable wallet, uint256 stakedAmount,
        uint256 totalTasks, uint256 successCount,
        bool isActive, bool isFeatured, uint256 listingFee,
        string memory endpointUrl, string memory description,
        uint256 registeredAt
    );
}

interface IAgentVaultForRep {
    function slashAgent(uint256 agentId, uint256 taskId, address user) external;
}

/**
 * @title ReputationEngine
 * @notice Calculates and stores agent reputation scores with auto-suspension
 */
contract ReputationEngine is Ownable {
    mapping(uint256 => uint256) public scores;
    mapping(uint256 => uint256[]) public scoreHistory;

    uint256 public constant SUSPENSION_THRESHOLD = 40;
    uint256 public constant REMOVAL_THRESHOLD = 20;

    IAgentRegistryForRep public registry;
    address public vault;

    // Authorized contracts
    mapping(address => bool) public authorized;

    event ScoreUpdated(uint256 indexed agentId, uint256 newScore);
    event AgentAutoSuspended(uint256 indexed agentId, uint256 score);
    event AgentAutoRemoved(uint256 indexed agentId, uint256 score);

    modifier onlyAuthorized() {
        require(authorized[msg.sender] || msg.sender == owner(), "Not authorized");
        _;
    }

    constructor(address _registry) Ownable(msg.sender) {
        registry = IAgentRegistryForRep(_registry);
    }

    function setVault(address _vault) external onlyOwner {
        vault = _vault;
    }

    function setAuthorized(address _contract, bool _status) external onlyOwner {
        authorized[_contract] = _status;
    }

    function updateScore(uint256 _agentId) external onlyAuthorized {
        (
            , , , , , , , uint256 totalTasks, uint256 successCount,
            , , , , ,
        ) = registry.agents(_agentId);

        uint256 newScore = 0;
        if (totalTasks > 0) {
            newScore = (successCount * 100) / totalTasks;
        }

        scores[_agentId] = newScore;
        scoreHistory[_agentId].push(newScore);

        // Keep only last 30 entries
        if (scoreHistory[_agentId].length > 30) {
            // Shift array (expensive but acceptable for testnet)
            uint256 len = scoreHistory[_agentId].length;
            for (uint256 i = 0; i < len - 1; i++) {
                scoreHistory[_agentId][i] = scoreHistory[_agentId][i + 1];
            }
            scoreHistory[_agentId].pop();
        }

        emit ScoreUpdated(_agentId, newScore);

        // Auto-suspension logic
        if (newScore > 0 && newScore < REMOVAL_THRESHOLD) {
            registry.suspendAgent(_agentId);
            emit AgentAutoRemoved(_agentId, newScore);
        } else if (newScore > 0 && newScore < SUSPENSION_THRESHOLD) {
            registry.suspendAgent(_agentId);
            emit AgentAutoSuspended(_agentId, newScore);
        }
    }

    function getScore(uint256 _agentId) external view returns (uint256) {
        return scores[_agentId];
    }

    function getScoreHistory(uint256 _agentId) external view returns (uint256[] memory) {
        uint256[] memory history = scoreHistory[_agentId];
        uint256 len = history.length;
        uint256 start = len > 30 ? len - 30 : 0;
        uint256 size = len - start;
        uint256[] memory result = new uint256[](size);
        for (uint256 i = 0; i < size; i++) {
            result[i] = history[start + i];
        }
        return result;
    }

    function isSuspended(uint256 _agentId) external view returns (bool) {
        return scores[_agentId] > 0 && scores[_agentId] < SUSPENSION_THRESHOLD;
    }

    // Manually set score for testing
    function setMockScore(uint256 _agentId, uint256 _score) external onlyOwner {
        scores[_agentId] = _score;
        scoreHistory[_agentId].push(_score);
        emit ScoreUpdated(_agentId, _score);
    }
}
