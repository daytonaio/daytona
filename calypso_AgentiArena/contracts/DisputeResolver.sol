// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";

interface ITaskLedgerForDispute {
    function getTaskUser(uint256 taskId) external view returns (address);
    function getTaskAgentId(uint256 taskId) external view returns (uint256);
    function setTaskStatus(uint256 taskId, uint8 status) external;
}

interface IAgentVaultForDispute {
    function slashAgent(uint256 agentId, uint256 taskId, address user) external;
    function releasePayment(uint256 taskId, uint256 agentId) external;
}

interface IAgentRegistryForDispute {
    function getAgentWallet(uint256 agentId) external view returns (address payable);
}

/**
 * @title DisputeResolver
 * @notice Handles disputed tasks fairly with evidence submission and resolution
 */
contract DisputeResolver is Ownable {
    enum DisputeStatus { OPEN, AGENT_RESPONDED, RESOLVED }
    enum DisputeOutcome { PENDING, USER_WINS, AGENT_WINS }

    struct Dispute {
        uint256 taskId;
        address user;
        uint256 agentId;
        string userEvidence;
        string agentResponse;
        DisputeStatus status;
        DisputeOutcome outcome;
        uint256 openedAt;
        uint256 resolvedAt;
    }

    mapping(uint256 => Dispute) public disputes;
    mapping(uint256 => bool) public hasDispute;
    uint256[] public openDisputeIds;

    ITaskLedgerForDispute public taskLedger;
    IAgentVaultForDispute public vault;
    IAgentRegistryForDispute public registry;

    event DisputeOpened(uint256 indexed taskId, address indexed user, uint256 indexed agentId);
    event DisputeResponse(uint256 indexed taskId, uint256 indexed agentId);
    event DisputeResolved(uint256 indexed taskId, DisputeOutcome outcome);

    constructor(
        address _taskLedger,
        address _vault,
        address _registry
    ) Ownable(msg.sender) {
        taskLedger = ITaskLedgerForDispute(_taskLedger);
        vault = IAgentVaultForDispute(_vault);
        registry = IAgentRegistryForDispute(_registry);
    }

    function openDispute(uint256 _taskId, string memory _evidence) external {
        address taskUser = taskLedger.getTaskUser(_taskId);
        require(msg.sender == taskUser, "Not task owner");
        require(!hasDispute[_taskId], "Dispute already exists");

        uint256 agentId = taskLedger.getTaskAgentId(_taskId);

        disputes[_taskId] = Dispute({
            taskId: _taskId,
            user: msg.sender,
            agentId: agentId,
            userEvidence: _evidence,
            agentResponse: "",
            status: DisputeStatus.OPEN,
            outcome: DisputeOutcome.PENDING,
            openedAt: block.timestamp,
            resolvedAt: 0
        });

        hasDispute[_taskId] = true;
        openDisputeIds.push(_taskId);

        emit DisputeOpened(_taskId, msg.sender, agentId);
    }

    function respondToDispute(uint256 _taskId, string memory _response) external {
        Dispute storage dispute = disputes[_taskId];
        require(hasDispute[_taskId], "No dispute found");
        require(dispute.status == DisputeStatus.OPEN, "Dispute not open");

        address payable agentWallet = registry.getAgentWallet(dispute.agentId);
        require(msg.sender == agentWallet, "Not agent owner");
        require(
            block.timestamp <= dispute.openedAt + 12 hours,
            "Response window expired"
        );

        dispute.agentResponse = _response;
        dispute.status = DisputeStatus.AGENT_RESPONDED;

        emit DisputeResponse(_taskId, dispute.agentId);
    }

    function resolveDispute(uint256 _taskId, bool _userWins) external onlyOwner {
        Dispute storage dispute = disputes[_taskId];
        require(hasDispute[_taskId], "No dispute found");
        require(
            dispute.status == DisputeStatus.OPEN ||
            dispute.status == DisputeStatus.AGENT_RESPONDED,
            "Already resolved"
        );

        dispute.resolvedAt = block.timestamp;
        dispute.status = DisputeStatus.RESOLVED;

        if (_userWins) {
            dispute.outcome = DisputeOutcome.USER_WINS;
            vault.slashAgent(dispute.agentId, _taskId, dispute.user);
        } else {
            dispute.outcome = DisputeOutcome.AGENT_WINS;
            vault.releasePayment(_taskId, dispute.agentId);
        }

        // Remove from open disputes
        _removeOpenDispute(_taskId);

        emit DisputeResolved(_taskId, dispute.outcome);
    }

    function getDispute(uint256 _taskId) external view returns (Dispute memory) {
        require(hasDispute[_taskId], "No dispute found");
        return disputes[_taskId];
    }

    function getOpenDisputes() external view returns (Dispute[] memory) {
        uint256 count = 0;
        for (uint256 i = 0; i < openDisputeIds.length; i++) {
            if (disputes[openDisputeIds[i]].status != DisputeStatus.RESOLVED) {
                count++;
            }
        }
        Dispute[] memory result = new Dispute[](count);
        uint256 idx = 0;
        for (uint256 i = 0; i < openDisputeIds.length; i++) {
            if (disputes[openDisputeIds[i]].status != DisputeStatus.RESOLVED) {
                result[idx] = disputes[openDisputeIds[i]];
                idx++;
            }
        }
        return result;
    }

    function _removeOpenDispute(uint256 _taskId) internal {
        for (uint256 i = 0; i < openDisputeIds.length; i++) {
            if (openDisputeIds[i] == _taskId) {
                openDisputeIds[i] = openDisputeIds[openDisputeIds.length - 1];
                openDisputeIds.pop();
                break;
            }
        }
    }
}
