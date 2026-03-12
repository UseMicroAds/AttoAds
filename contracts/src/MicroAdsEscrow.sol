// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20} from "./interfaces/IERC20.sol";

/// @title MicroAdsEscrow
/// @notice Holds USDC deposits for advertising campaigns and releases funds
///         to commenters once the backend verification worker confirms the
///         comment edit. Only the designated operator can trigger releases.
contract MicroAdsEscrow {
    IERC20 public immutable usdc;
    address public operator;
    address public owner;

    struct CampaignEscrow {
        address advertiser;
        uint256 deposited;
        uint256 released;
        bool refunded;
    }

    mapping(bytes32 => CampaignEscrow) public campaigns;

    event Deposited(bytes32 indexed campaignId, address indexed advertiser, uint256 amount);
    event Released(bytes32 indexed dealId, bytes32 indexed campaignId, address indexed commenter, uint256 amount);
    event Refunded(bytes32 indexed campaignId, address indexed advertiser, uint256 amount);
    event OperatorUpdated(address indexed oldOperator, address indexed newOperator);

    error NotOwner();
    error NotOperator();
    error NotAdvertiser();
    error ZeroAmount();
    error InsufficientEscrow();
    error AlreadyRefunded();
    error TransferFailed();

    modifier onlyOwner() {
        if (msg.sender != owner) revert NotOwner();
        _;
    }

    modifier onlyOperator() {
        if (msg.sender != operator) revert NotOperator();
        _;
    }

    constructor(address _usdc, address _operator) {
        usdc = IERC20(_usdc);
        operator = _operator;
        owner = msg.sender;
    }

    /// @notice Advertiser deposits USDC for a campaign. Requires prior ERC-20 approval.
    function deposit(bytes32 campaignId, uint256 amount) external {
        if (amount == 0) revert ZeroAmount();

        CampaignEscrow storage c = campaigns[campaignId];
        if (c.advertiser == address(0)) {
            c.advertiser = msg.sender;
        }

        bool ok = usdc.transferFrom(msg.sender, address(this), amount);
        if (!ok) revert TransferFailed();

        c.deposited += amount;

        emit Deposited(campaignId, msg.sender, amount);
    }

    /// @notice Operator releases funds to a commenter after verification.
    function release(
        bytes32 dealId,
        bytes32 campaignId,
        address commenter,
        uint256 amount
    ) external onlyOperator {
        if (amount == 0) revert ZeroAmount();

        CampaignEscrow storage c = campaigns[campaignId];
        uint256 available = c.deposited - c.released;
        if (amount > available) revert InsufficientEscrow();

        c.released += amount;

        bool ok = usdc.transfer(commenter, amount);
        if (!ok) revert TransferFailed();

        emit Released(dealId, campaignId, commenter, amount);
    }

    /// @notice Advertiser reclaims remaining un-released funds.
    function refund(bytes32 campaignId) external {
        CampaignEscrow storage c = campaigns[campaignId];
        if (msg.sender != c.advertiser) revert NotAdvertiser();
        if (c.refunded) revert AlreadyRefunded();

        uint256 remaining = c.deposited - c.released;
        if (remaining == 0) revert ZeroAmount();

        c.refunded = true;

        bool ok = usdc.transfer(c.advertiser, remaining);
        if (!ok) revert TransferFailed();

        emit Refunded(campaignId, c.advertiser, remaining);
    }

    function setOperator(address _operator) external onlyOwner {
        emit OperatorUpdated(operator, _operator);
        operator = _operator;
    }

    function getAvailableBalance(bytes32 campaignId) external view returns (uint256) {
        CampaignEscrow storage c = campaigns[campaignId];
        if (c.refunded) return 0;
        return c.deposited - c.released;
    }
}
