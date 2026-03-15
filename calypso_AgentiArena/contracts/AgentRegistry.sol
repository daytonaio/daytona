// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title AgentRegistry
 * @notice Stores all agent listings on-chain
 */
contract AgentRegistry is Ownable, ReentrancyGuard {
    struct Agent {
        uint256 id;
        string name;
        string category;
        string framework;
        uint256 pricePerCall;
        address payable wallet;
        uint256 stakedAmount;
        uint256 totalTasks;
        uint256 successCount;
        bool isActive;
        bool isFeatured;
        uint256 listingFee;
        string endpointUrl;
        string description;
        uint256 registeredAt;
    }

    mapping(uint256 => Agent) public agents;
    uint256 public agentCount;
    uint256 public constant MIN_STAKE = 100 * 10 ** 18;
    uint256 public constant LISTING_FEE = 5 * 10 ** 18;
    address public treasury;
    IERC20 public hlusd;
    address public vault;

    // Mapping from wallet to agent IDs
    mapping(address => uint256[]) public ownerAgents;

    event AgentRegistered(uint256 indexed agentId, address indexed wallet, string name);
    event AgentStatusUpdated(uint256 indexed agentId, bool isActive);
    event AgentDelisted(uint256 indexed agentId);
    event AgentStatsUpdated(uint256 indexed agentId, uint256 totalTasks, uint256 successCount);

    constructor(address _hlusd, address _treasury) Ownable(msg.sender) {
        hlusd = IERC20(_hlusd);
        treasury = _treasury;
    }

    function setVault(address _vault) external onlyOwner {
        vault = _vault;
    }

    function registerAgent(
        string memory _name,
        string memory _category,
        string memory _framework,
        uint256 _pricePerCall,
        string memory _endpointUrl,
        string memory _description,
        uint256 _stakeAmount
    ) external nonReentrant returns (uint256) {
        require(_stakeAmount >= MIN_STAKE, "Stake below minimum");
        require(
            hlusd.allowance(msg.sender, address(this)) >= _stakeAmount + LISTING_FEE,
            "Insufficient allowance"
        );

        // Transfer listing fee to treasury
        require(hlusd.transferFrom(msg.sender, treasury, LISTING_FEE), "Listing fee transfer failed");

        // Transfer stake to vault
        require(hlusd.transferFrom(msg.sender, vault, _stakeAmount), "Stake transfer failed");

        agentCount++;
        agents[agentCount] = Agent({
            id: agentCount,
            name: _name,
            category: _category,
            framework: _framework,
            pricePerCall: _pricePerCall,
            wallet: payable(msg.sender),
            stakedAmount: _stakeAmount,
            totalTasks: 0,
            successCount: 0,
            isActive: true,
            isFeatured: false,
            listingFee: LISTING_FEE,
            endpointUrl: _endpointUrl,
            description: _description,
            registeredAt: block.timestamp
        });

        ownerAgents[msg.sender].push(agentCount);

        emit AgentRegistered(agentCount, msg.sender, _name);
        return agentCount;
    }

    function getAgent(uint256 _agentId) external view returns (Agent memory) {
        require(_agentId > 0 && _agentId <= agentCount, "Agent not found");
        return agents[_agentId];
    }

    function getAllAgents() external view returns (Agent[] memory) {
        Agent[] memory allAgents = new Agent[](agentCount);
        for (uint256 i = 1; i <= agentCount; i++) {
            allAgents[i - 1] = agents[i];
        }
        return allAgents;
    }

    function getAgentsByCategory(string memory _category) external view returns (Agent[] memory) {
        uint256 count = 0;
        for (uint256 i = 1; i <= agentCount; i++) {
            if (keccak256(bytes(agents[i].category)) == keccak256(bytes(_category))) {
                count++;
            }
        }
        Agent[] memory result = new Agent[](count);
        uint256 idx = 0;
        for (uint256 i = 1; i <= agentCount; i++) {
            if (keccak256(bytes(agents[i].category)) == keccak256(bytes(_category))) {
                result[idx] = agents[i];
                idx++;
            }
        }
        return result;
    }

    function getTopAgents(uint256 _limit) external view returns (Agent[] memory) {
        uint256 len = _limit > agentCount ? agentCount : _limit;
        Agent[] memory sorted = new Agent[](agentCount);
        for (uint256 i = 1; i <= agentCount; i++) {
            sorted[i - 1] = agents[i];
        }
        // Simple bubble sort by success rate
        for (uint256 i = 0; i < sorted.length; i++) {
            for (uint256 j = i + 1; j < sorted.length; j++) {
                uint256 rateI = sorted[i].totalTasks > 0
                    ? (sorted[i].successCount * 10000) / sorted[i].totalTasks
                    : 0;
                uint256 rateJ = sorted[j].totalTasks > 0
                    ? (sorted[j].successCount * 10000) / sorted[j].totalTasks
                    : 0;
                if (rateJ > rateI) {
                    Agent memory tmp = sorted[i];
                    sorted[i] = sorted[j];
                    sorted[j] = tmp;
                }
            }
        }
        Agent[] memory top = new Agent[](len);
        for (uint256 i = 0; i < len; i++) {
            top[i] = sorted[i];
        }
        return top;
    }

    function updateAgentStatus(uint256 _agentId, bool _isActive) external {
        Agent storage agent = agents[_agentId];
        require(agent.id > 0, "Agent not found");
        require(
            msg.sender == agent.wallet || msg.sender == owner(),
            "Not authorized"
        );
        agent.isActive = _isActive;
        emit AgentStatusUpdated(_agentId, _isActive);
    }

    function delistAgent(uint256 _agentId) external {
        Agent storage agent = agents[_agentId];
        require(agent.id > 0, "Agent not found");
        require(msg.sender == agent.wallet, "Not agent owner");
        agent.isActive = false;
        emit AgentDelisted(_agentId);
    }

    // --- Functions called by other contracts ---

    function incrementTaskCount(uint256 _agentId) external {
        agents[_agentId].totalTasks++;
    }

    function incrementSuccessCount(uint256 _agentId) external {
        agents[_agentId].successCount++;
    }

    function suspendAgent(uint256 _agentId) external {
        agents[_agentId].isActive = false;
        emit AgentStatusUpdated(_agentId, false);
    }

    function getAgentWallet(uint256 _agentId) external view returns (address payable) {
        return agents[_agentId].wallet;
    }

    function getAgentStake(uint256 _agentId) external view returns (uint256) {
        return agents[_agentId].stakedAmount;
    }

    function reduceStake(uint256 _agentId, uint256 _amount) external {
        agents[_agentId].stakedAmount -= _amount;
    }

    function addStake(uint256 _agentId, uint256 _amount) external {
        agents[_agentId].stakedAmount += _amount;
    }

    // --- Testnet only: set mock stats ---

    function setMockStats(uint256 _agentId, uint256 _totalTasks, uint256 _successCount) external onlyOwner {
        agents[_agentId].totalTasks = _totalTasks;
        agents[_agentId].successCount = _successCount;
        emit AgentStatsUpdated(_agentId, _totalTasks, _successCount);
    }

    function setFeatured(uint256 _agentId, bool _featured) external onlyOwner {
        agents[_agentId].isFeatured = _featured;
    }
}
