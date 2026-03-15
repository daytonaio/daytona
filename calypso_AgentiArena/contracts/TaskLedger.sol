// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";

interface IAgentRegistry {
    function incrementTaskCount(uint256 agentId) external;
    function incrementSuccessCount(uint256 agentId) external;
    function getAgentWallet(uint256 agentId) external view returns (address payable);
}

interface IAgentVault {
    function releasePayment(uint256 taskId, uint256 agentId) external;
    function refundUser(uint256 taskId, address user) external;
    function payForTask(uint256 agentId, uint256 taskId, uint256 amount, address user, bool isArena) external;
}

/**
 * @title TaskLedger
 * @notice Records every task permanently on-chain — the reputation proof system
 */
contract TaskLedger is Ownable {
    enum TaskStatus { PENDING, SUCCESS, FAILED, DISPUTED }

    struct Task {
        uint256 id;
        uint256 agentId;
        address user;
        string taskType;
        TaskStatus status;
        uint256 paidAmount;
        bool isArena;
        uint256 createdAt;
        uint256 completedAt;
        bytes32 inputHash;
        bytes32 outputHash;
        string resultSummary;
    }

    mapping(uint256 => Task) public tasks;
    mapping(address => uint256[]) public userTasks;
    mapping(uint256 => uint256[]) public agentTasks;
    uint256 public taskCount;

    IAgentRegistry public registry;
    IAgentVault public vault;

    // Authorized contracts
    mapping(address => bool) public authorized;

    event TaskCreated(uint256 indexed taskId, uint256 indexed agentId, address indexed user);
    event TaskCompleted(uint256 indexed taskId, uint256 indexed agentId, bytes32 outputHash);
    event TaskFailed(uint256 indexed taskId, uint256 indexed agentId, string reason);
    event TaskDisputed(uint256 indexed taskId, address indexed user);

    modifier onlyAuthorized() {
        require(authorized[msg.sender] || msg.sender == owner(), "Not authorized");
        _;
    }

    constructor(address _registry, address _vault) Ownable(msg.sender) {
        registry = IAgentRegistry(_registry);
        vault = IAgentVault(_vault);
    }

    function setAuthorized(address _contract, bool _status) external onlyOwner {
        authorized[_contract] = _status;
    }

    function setVault(address _vault) external onlyOwner {
        vault = IAgentVault(_vault);
    }

    function createTask(
        uint256 _agentId,
        address _user,
        string memory _taskType,
        uint256 _paidAmount,
        bool _isArena,
        bytes32 _inputHash
    ) external onlyAuthorized returns (uint256) {
        taskCount++;
        tasks[taskCount] = Task({
            id: taskCount,
            agentId: _agentId,
            user: _user,
            taskType: _taskType,
            status: TaskStatus.PENDING,
            paidAmount: _paidAmount,
            isArena: _isArena,
            createdAt: block.timestamp,
            completedAt: 0,
            inputHash: _inputHash,
            outputHash: bytes32(0),
            resultSummary: ""
        });

        userTasks[_user].push(taskCount);
        agentTasks[_agentId].push(taskCount);
        registry.incrementTaskCount(_agentId);

        emit TaskCreated(taskCount, _agentId, _user);
        return taskCount;
    }

    function completeTask(
        uint256 _taskId,
        bytes32 _outputHash,
        string memory _resultSummary
    ) external onlyAuthorized {
        Task storage task = tasks[_taskId];
        require(task.id > 0, "Task not found");
        require(task.status == TaskStatus.PENDING, "Task not pending");

        task.status = TaskStatus.SUCCESS;
        task.completedAt = block.timestamp;
        task.outputHash = _outputHash;
        task.resultSummary = _resultSummary;

        registry.incrementSuccessCount(task.agentId);
        vault.releasePayment(_taskId, task.agentId);

        emit TaskCompleted(_taskId, task.agentId, _outputHash);
    }

    function failTask(uint256 _taskId, string memory _reason) external onlyAuthorized {
        Task storage task = tasks[_taskId];
        require(task.id > 0, "Task not found");
        require(task.status == TaskStatus.PENDING, "Task not pending");

        task.status = TaskStatus.FAILED;
        task.completedAt = block.timestamp;
        task.resultSummary = _reason;

        vault.refundUser(_taskId, task.user);

        emit TaskFailed(_taskId, task.agentId, _reason);
    }

    function disputeTask(uint256 _taskId) external {
        Task storage task = tasks[_taskId];
        require(task.id > 0, "Task not found");
        require(msg.sender == task.user, "Not task owner");
        require(
            task.status == TaskStatus.SUCCESS,
            "Can only dispute completed tasks"
        );
        require(
            block.timestamp <= task.completedAt + 24 hours,
            "Dispute window expired"
        );

        task.status = TaskStatus.DISPUTED;

        emit TaskDisputed(_taskId, msg.sender);
    }

    function getTask(uint256 _taskId) external view returns (Task memory) {
        require(tasks[_taskId].id > 0, "Task not found");
        return tasks[_taskId];
    }

    function getUserTasks(address _user) external view returns (Task[] memory) {
        uint256[] memory ids = userTasks[_user];
        Task[] memory result = new Task[](ids.length);
        for (uint256 i = 0; i < ids.length; i++) {
            result[i] = tasks[ids[i]];
        }
        return result;
    }

    function getAgentTasks(uint256 _agentId) external view returns (Task[] memory) {
        uint256[] memory ids = agentTasks[_agentId];
        Task[] memory result = new Task[](ids.length);
        for (uint256 i = 0; i < ids.length; i++) {
            result[i] = tasks[ids[i]];
        }
        return result;
    }

    function getAgentSuccessRate(uint256 _agentId) external view returns (uint256) {
        uint256[] memory ids = agentTasks[_agentId];
        if (ids.length == 0) return 0;
        uint256 successCount = 0;
        for (uint256 i = 0; i < ids.length; i++) {
            if (tasks[ids[i]].status == TaskStatus.SUCCESS) {
                successCount++;
            }
        }
        return (successCount * 100) / ids.length;
    }

    // Allow setting task status for dispute resolution
    function setTaskStatus(uint256 _taskId, TaskStatus _status) external onlyAuthorized {
        tasks[_taskId].status = _status;
    }

    function getTaskUser(uint256 _taskId) external view returns (address) {
        return tasks[_taskId].user;
    }

    function getTaskAgentId(uint256 _taskId) external view returns (uint256) {
        return tasks[_taskId].agentId;
    }
}
