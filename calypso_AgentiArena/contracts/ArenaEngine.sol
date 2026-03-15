// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

interface IAgentRegistryForArena {
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

interface IAgentVaultForArena {
    function payForTask(uint256 agentId, uint256 taskId, uint256 amount, address user, bool isArena) external;
}

interface ITaskLedgerForArena {
    function createTask(
        uint256 agentId, address user, string memory taskType,
        uint256 paidAmount, bool isArena, bytes32 inputHash
    ) external returns (uint256);
}

/**
 * @title ArenaEngine
 * @notice Manages reverse auction arena battles
 */
contract ArenaEngine is Ownable, ReentrancyGuard {
    enum ArenaStatus { OPEN, SELECTING, CLOSED, CANCELLED }

    struct Arena {
        uint256 id;
        address user;
        string taskDescription;
        string category;
        uint256 maxBudget;
        uint256 deadline;
        ArenaStatus status;
        uint256 winnerAgentId;
        uint256 winningBid;
        uint256 createdAt;
    }

    struct Bid {
        uint256 agentId;
        uint256 amount;
        uint256 submittedAt;
        bool selected;
    }

    mapping(uint256 => Arena) public arenas;
    mapping(uint256 => Bid[]) public arenaBids;
    mapping(uint256 => mapping(uint256 => bool)) public agentHasBid; // arenaId => agentId => hasBid
    uint256 public arenaCount;
    uint256 public constant ARENA_FEE_PREMIUM = 2;

    IERC20 public hlusd;
    IAgentRegistryForArena public registry;
    address public vault;
    address public taskLedger;

    event ArenaCreated(uint256 indexed arenaId, address indexed user, string category, uint256 maxBudget);
    event BidSubmitted(uint256 indexed arenaId, uint256 indexed agentId, uint256 bidAmount);
    event WinnerSelected(uint256 indexed arenaId, uint256 indexed agentId, uint256 bidAmount);
    event ArenaCancelled(uint256 indexed arenaId);

    constructor(
        address _hlusd,
        address _registry,
        address _vault,
        address _taskLedger
    ) Ownable(msg.sender) {
        hlusd = IERC20(_hlusd);
        registry = IAgentRegistryForArena(_registry);
        vault = _vault;
        taskLedger = _taskLedger;
    }

    function createArena(
        string memory _taskDescription,
        string memory _category,
        uint256 _maxBudget,
        uint256 _durationSeconds
    ) external nonReentrant returns (uint256) {
        require(_maxBudget > 0, "Budget must be > 0");
        require(_durationSeconds >= 30, "Min duration 30s");

        // User deposits maxBudget into escrow
        require(
            hlusd.transferFrom(msg.sender, address(this), _maxBudget),
            "Budget transfer failed"
        );

        arenaCount++;
        arenas[arenaCount] = Arena({
            id: arenaCount,
            user: msg.sender,
            taskDescription: _taskDescription,
            category: _category,
            maxBudget: _maxBudget,
            deadline: block.timestamp + _durationSeconds,
            status: ArenaStatus.OPEN,
            winnerAgentId: 0,
            winningBid: 0,
            createdAt: block.timestamp
        });

        emit ArenaCreated(arenaCount, msg.sender, _category, _maxBudget);
        return arenaCount;
    }

    function submitBid(uint256 _arenaId, uint256 _agentId, uint256 _bidAmount) external {
        Arena storage arena = arenas[_arenaId];
        require(arena.id > 0, "Arena not found");
        require(arena.status == ArenaStatus.OPEN, "Arena not open");
        require(block.timestamp <= arena.deadline, "Arena deadline passed");
        require(_bidAmount <= arena.maxBudget, "Bid exceeds budget");
        require(!agentHasBid[_arenaId][_agentId], "Agent already bid");

        // Check agent is active and has sufficient reputation
        (
            uint256 id, , , , , , , , uint256 successCount,
            bool isActive, , , , ,
        ) = registry.agents(_agentId);
        require(id > 0, "Agent not found");
        require(isActive, "Agent not active");

        agentHasBid[_arenaId][_agentId] = true;

        arenaBids[_arenaId].push(Bid({
            agentId: _agentId,
            amount: _bidAmount,
            submittedAt: block.timestamp,
            selected: false
        }));

        emit BidSubmitted(_arenaId, _agentId, _bidAmount);
    }

    function selectWinner(uint256 _arenaId, uint256 _agentId) external nonReentrant {
        Arena storage arena = arenas[_arenaId];
        require(arena.id > 0, "Arena not found");
        require(msg.sender == arena.user, "Not arena creator");
        require(
            arena.status == ArenaStatus.OPEN || arena.status == ArenaStatus.SELECTING,
            "Arena not selectable"
        );

        // Find the bid
        Bid[] storage bids = arenaBids[_arenaId];
        bool found = false;
        uint256 bidAmount = 0;
        for (uint256 i = 0; i < bids.length; i++) {
            if (bids[i].agentId == _agentId) {
                bids[i].selected = true;
                bidAmount = bids[i].amount;
                found = true;
                break;
            }
        }
        require(found, "Agent has no bid");

        arena.status = ArenaStatus.CLOSED;
        arena.winnerAgentId = _agentId;
        arena.winningBid = bidAmount;

        // Refund difference to user
        uint256 refund = arena.maxBudget - bidAmount;
        if (refund > 0) {
            require(hlusd.transfer(arena.user, refund), "Refund failed");
        }

        // Transfer winning bid to vault for payment processing
        require(hlusd.transfer(vault, bidAmount), "Transfer to vault failed");

        emit WinnerSelected(_arenaId, _agentId, bidAmount);
    }

    function cancelArena(uint256 _arenaId) external nonReentrant {
        Arena storage arena = arenas[_arenaId];
        require(arena.id > 0, "Arena not found");
        require(
            msg.sender == arena.user || block.timestamp > arena.deadline,
            "Not authorized"
        );
        require(arena.status == ArenaStatus.OPEN, "Arena not open");

        arena.status = ArenaStatus.CANCELLED;
        require(hlusd.transfer(arena.user, arena.maxBudget), "Refund failed");

        emit ArenaCancelled(_arenaId);
    }

    function getArena(uint256 _arenaId) external view returns (Arena memory) {
        require(arenas[_arenaId].id > 0, "Arena not found");
        return arenas[_arenaId];
    }

    function getArenaBids(uint256 _arenaId) external view returns (Bid[] memory) {
        return arenaBids[_arenaId];
    }

    function getOpenArenas() external view returns (Arena[] memory) {
        uint256 count = 0;
        for (uint256 i = 1; i <= arenaCount; i++) {
            if (arenas[i].status == ArenaStatus.OPEN) count++;
        }
        Arena[] memory result = new Arena[](count);
        uint256 idx = 0;
        for (uint256 i = 1; i <= arenaCount; i++) {
            if (arenas[i].status == ArenaStatus.OPEN) {
                result[idx] = arenas[i];
                idx++;
            }
        }
        return result;
    }

    function getArenaBidCount(uint256 _arenaId) external view returns (uint256) {
        return arenaBids[_arenaId].length;
    }
}
