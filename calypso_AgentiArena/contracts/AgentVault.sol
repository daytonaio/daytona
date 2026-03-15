// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title AgentVault
 * @notice Holds all payments and stake in escrow with payment splitting
 */
contract AgentVault is Ownable, ReentrancyGuard {
    IERC20 public hlusd;
    address public treasury;

    // Payment split constants (percentages)
    uint256 public constant AGENT_SHARE = 85;
    uint256 public constant PROTOCOL_FEE = 10;
    uint256 public constant REPUTATION_POOL_FEE = 5;
    uint256 public constant ARENA_PREMIUM = 2;

    // Slash distribution
    uint256 public constant SLASH_TO_USER = 70;
    uint256 public constant SLASH_TO_TREASURY = 20;
    uint256 public constant SLASH_BURN = 10;
    address public constant BURN_ADDRESS = 0x000000000000000000000000000000000000dEaD;

    mapping(uint256 => uint256) public agentStakes;
    mapping(uint256 => uint256) public agentEarnings;
    mapping(uint256 => uint256) public taskEscrow;
    mapping(uint256 => address) public taskPayer;
    mapping(uint256 => uint256) public taskAgentId;
    uint256 public treasuryBalance;
    uint256 public reputationPool;

    // Authorized contracts
    mapping(address => bool) public authorized;

    event TaskPaid(uint256 indexed taskId, uint256 indexed agentId, uint256 amount, address user);
    event PaymentReleased(uint256 indexed taskId, uint256 indexed agentId, uint256 amount);
    event UserRefunded(uint256 indexed taskId, address indexed user, uint256 amount);
    event AgentSlashed(uint256 indexed agentId, uint256 amount, address indexed user);
    event StakeDeposited(uint256 indexed agentId, uint256 amount);
    event StakeWithdrawn(uint256 indexed agentId, uint256 amount);
    event EarningsWithdrawn(uint256 indexed agentId, uint256 amount);
    event TreasuryWithdrawn(address indexed to, uint256 amount);

    modifier onlyAuthorized() {
        require(authorized[msg.sender] || msg.sender == owner(), "Not authorized");
        _;
    }

    constructor(address _hlusd, address _treasury) Ownable(msg.sender) {
        hlusd = IERC20(_hlusd);
        treasury = _treasury;
    }

    function setAuthorized(address _contract, bool _status) external onlyOwner {
        authorized[_contract] = _status;
    }

    function payForTask(
        uint256 _agentId,
        uint256 _taskId,
        uint256 _amount,
        address _user,
        bool _isArena
    ) external onlyAuthorized nonReentrant {
        require(_amount > 0, "Amount must be > 0");
        require(taskEscrow[_taskId] == 0, "Task already paid");

        require(hlusd.transferFrom(_user, address(this), _amount), "Payment transfer failed");

        taskEscrow[_taskId] = _amount;
        taskPayer[_taskId] = _user;
        taskAgentId[_taskId] = _agentId;

        emit TaskPaid(_taskId, _agentId, _amount, _user);
    }

    function releasePayment(uint256 _taskId, uint256 _agentId) external onlyAuthorized nonReentrant {
        uint256 amount = taskEscrow[_taskId];
        require(amount > 0, "No escrowed amount");

        uint256 agentAmount = (amount * AGENT_SHARE) / 100;
        uint256 protocolAmount = (amount * PROTOCOL_FEE) / 100;
        uint256 reputationAmount = (amount * REPUTATION_POOL_FEE) / 100;

        agentEarnings[_agentId] += agentAmount;
        treasuryBalance += protocolAmount;
        reputationPool += reputationAmount;

        taskEscrow[_taskId] = 0;

        emit PaymentReleased(_taskId, _agentId, amount);
    }

    function refundUser(uint256 _taskId, address _user) external onlyAuthorized nonReentrant {
        uint256 amount = taskEscrow[_taskId];
        require(amount > 0, "No escrowed amount");

        taskEscrow[_taskId] = 0;
        require(hlusd.transfer(_user, amount), "Refund failed");

        emit UserRefunded(_taskId, _user, amount);
    }

    function slashAgent(uint256 _agentId, uint256 _taskId, address _user) external onlyAuthorized nonReentrant {
        uint256 stake = agentStakes[_agentId];
        require(stake > 0, "No stake to slash");

        uint256 slashAmount = (stake * 20) / 100; // 20% of stake
        uint256 userCompensation = (slashAmount * SLASH_TO_USER) / 100;
        uint256 treasuryShareAmount = (slashAmount * SLASH_TO_TREASURY) / 100;
        uint256 burnAmount = (slashAmount * SLASH_BURN) / 100;

        agentStakes[_agentId] -= slashAmount;

        require(hlusd.transfer(_user, userCompensation), "User compensation failed");
        treasuryBalance += treasuryShareAmount;
        require(hlusd.transfer(BURN_ADDRESS, burnAmount), "Burn failed");

        emit AgentSlashed(_agentId, slashAmount, _user);
    }

    function depositStake(uint256 _agentId, uint256 _amount) external nonReentrant {
        require(_amount > 0, "Amount must be > 0");
        require(hlusd.transferFrom(msg.sender, address(this), _amount), "Transfer failed");
        agentStakes[_agentId] += _amount;
        emit StakeDeposited(_agentId, _amount);
    }

    function withdrawStake(uint256 _agentId, uint256 _amount) external nonReentrant {
        require(agentStakes[_agentId] >= _amount, "Insufficient stake");
        agentStakes[_agentId] -= _amount;
        require(hlusd.transfer(msg.sender, _amount), "Transfer failed");
        emit StakeWithdrawn(_agentId, _amount);
    }

    function withdrawAgentEarnings(uint256 _agentId) external nonReentrant {
        uint256 amount = agentEarnings[_agentId];
        require(amount > 0, "No earnings");
        agentEarnings[_agentId] = 0;
        require(hlusd.transfer(msg.sender, amount), "Transfer failed");
        emit EarningsWithdrawn(_agentId, amount);
    }

    function withdrawTreasury(address _to, uint256 _amount) external onlyOwner nonReentrant {
        require(_amount <= treasuryBalance, "Insufficient treasury");
        treasuryBalance -= _amount;
        require(hlusd.transfer(_to, _amount), "Transfer failed");
        emit TreasuryWithdrawn(_to, _amount);
    }

    // Initialize stake when agent registers (called by registry)
    function initStake(uint256 _agentId, uint256 _amount) external onlyAuthorized {
        agentStakes[_agentId] += _amount;
    }
}
