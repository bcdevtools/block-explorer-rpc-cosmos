package types

import (
	"github.com/ethereum/go-ethereum/common"
)

//goland:noinspection GoSnakeCaseUsage,SpellCheckingInspection
var (
	EvmEvent_Erc20_Erc721_Transfer = common.HexToHash(
		"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
	) // Keccak-256 Transfer(address,address,uint256)
	EvmEvent_Erc20_Erc721_Approval = common.HexToHash(
		"0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
	) // Keccak-256 Approval(address,address,uint256)
	EvmEvent_Erc721_Erc1155_ApprovalForAll = common.HexToHash(
		"0x17307eab39ab6107e8899845ad3d59bd9653f200f220920489ca2b5937696c31",
	) // Keccak-256 ApprovalForAll(address,address,bool)
	EvmEvent_Erc1155_TransferSingle = common.HexToHash(
		"0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62",
	) // Keccak-256 TransferSingle(address,address,address,uint256,uint256)
	EvmEvent_Erc1155_TransferBatch = common.HexToHash(
		"0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb",
	) // Keccak-256 TransferBatch(address,address,address,uint256[],uint256[])
	EvmEvent_WDeposit = common.HexToHash(
		"0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c",
	) // Deposit(address indexed user,uint256 amount)
	EvmEvent_WWithdraw = common.HexToHash(
		"0x7fcf532c15f0a6db0bd6d0e038bea71d30d808c7d98cb3bf7268a95bf5081b65",
	) // Withdrawal(address indexed src,uint256 wad)
)
