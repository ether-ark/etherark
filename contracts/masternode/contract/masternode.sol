pragma solidity ^0.5.4;

contract Masternode {

    uint public constant nodeCost = 10000 * 10**18;
    uint public constant baseCost = 10**18;
    uint public constant minBlockTimeout = 800;

    bytes8 public lastId;
    bytes8 public lastOnlineId;
    uint public countTotalNode;
    uint public countOnlineNode;

    struct node {
        bytes32 id1;
        bytes32 id2;
        bytes8 preId;
        bytes8 nextId;
        bytes8 preOnlineId;
        bytes8 nextOnlineId;
        address coinbase;
        uint blockRegister;
        uint blockLastPing;
        uint blockOnline;
        uint blockOnlineAcc;
    }

    mapping (bytes8 => node) public nodes;
    mapping (address => bytes8) nodeAddressToId;
    mapping (address => bytes8[]) public idsOf;

    event join(bytes8 id, address addr);
    event quit(bytes8 id, address addr);

    function register(bytes32 id1, bytes32 id2) public payable{
        register2(id1, id2, msg.sender);
    }

    function register2(bytes32 id1, bytes32 id2, address owner) public payable{
        bytes8 id = bytes8(id1);
        require(
            bytes8(0) != id &&
            bytes32(0) != id2 &&
            bytes32(0) == nodes[id].id1 &&
            msg.value == nodeCost
        );
        bytes32[2] memory input;
        address payable[1]  memory output;
        input[0] = id1;
        input[1] = id2;
        assembly {
            if iszero(call(not(0), 0x0b, 0, input, 128, output, 32)) {
                revert(0, 0)
            }
        }
        address payable account = output[0];
        require(account != address(0));
        nodeAddressToId[account] = id;
        nodes[id] = node(
            id1,id2,
            lastId,bytes8(0),
            bytes8(0),bytes8(0),
            owner,
            block.number,0,0,0
        );
        if(lastId != bytes8(0)){
            nodes[lastId].nextId = id;
        }
        lastId = id;
        idsOf[owner].push(id);
        countTotalNode += 1;
        account.transfer(baseCost);
        emit join(id, owner);
    }

    function() external {
        bytes8 id = nodeAddressToId[msg.sender];
        if (id != bytes8(0)){
            // ping
            if(0 == nodes[id].blockOnline){
                nodes[id].blockOnline = 1;
                countOnlineNode += 1;
                if(lastOnlineId != bytes8(0)){
                    nodes[lastOnlineId].nextOnlineId = id;
                }
                nodes[id].preOnlineId = lastOnlineId;
                nodes[id].nextOnlineId = bytes8(0);
                lastOnlineId = id;
            }else if(nodes[id].blockLastPing > 0){
                uint blockGap = block.number - nodes[id].blockLastPing;
                if(blockGap > minBlockTimeout){
                    nodes[id].blockOnline = 1;
                }else{
                    nodes[id].blockOnline += blockGap;
                    nodes[id].blockOnlineAcc += blockGap;
                }
            }
            nodes[id].blockLastPing = block.number;
            fix(nodes[id].preOnlineId);
            fix(nodes[id].nextOnlineId);
        }else if(idsOf[msg.sender].length > 0){
            uint index = idsOf[msg.sender].length -1;
            id = idsOf[msg.sender][index];
            bytes32 id1 = nodes[id].id1;
            require(
                bytes8(0) != id &&
                bytes32(0) != id1
            );
            offline(id);
            bytes32[2] memory input;
            address payable[1]  memory output;
            input[0] = id1;
            input[1] = nodes[id].id2;
            assembly {
                if iszero(call(not(0), 0x0b, 0, input, 128, output, 32)) {
                  revert(0, 0)
                }
            }
            address payable account = output[0];
            nodeAddressToId[account] = bytes8(0);

            bytes8 preId = nodes[id].preId;
            bytes8 nextId = nodes[id].nextId;
            if(preId != bytes8(0)){
                nodes[preId].nextId = nextId;
            }
            if(nextId != bytes8(0)){
                nodes[nextId].preId = preId;
            }else{
                lastId = preId;
            }
            bool notGenesisNode = nodes[id].blockRegister > 0;
            nodes[id] = node(
                bytes32(0),
                bytes32(0),
                bytes8(0),
                bytes8(0),
                bytes8(0),
                bytes8(0),
                address(0),
                uint(0),
                uint(0),
                uint(0),
                uint(0)
            );
            idsOf[msg.sender][index] = bytes8(0);
            idsOf[msg.sender].length = index;
            countTotalNode -= 1;
            emit quit(id, msg.sender);
            if(notGenesisNode){
                msg.sender.transfer(nodeCost - baseCost);
            }
        }
    }

    function fix(bytes8 id) internal {
        if (id != bytes8(0) && nodes[id].id1 != bytes32(0)){
            if((block.number - nodes[id].blockLastPing) > minBlockTimeout){
                offline(id);
            }
        }
    }

    function offline(bytes8 id) internal {
        if (nodes[id].blockOnline > 0){
            countOnlineNode -= 1;
            nodes[id].blockOnline = 0;
            bytes8 preOnlineId = nodes[id].preOnlineId;
            bytes8 nextOnlineId = nodes[id].nextOnlineId;
            if(preOnlineId != bytes8(0)){
                nodes[preOnlineId].nextOnlineId = nextOnlineId;
                nodes[id].preOnlineId = bytes8(0);
            }
            if(nextOnlineId != bytes8(0)){
                nodes[nextOnlineId].preOnlineId = preOnlineId;
                nodes[id].nextOnlineId = bytes8(0);
            }else{
                lastOnlineId = preOnlineId;
            }
        }
    }

    function getInfo(address addr) view public returns (
        uint lockedBalance,
        uint releasedReward,
        uint totalNodes,
        uint onlineNodes,
        uint myNodes
    )
    {
        lockedBalance = address(this).balance / (10**18);
        releasedReward = block.number * 48 / 10;
        totalNodes = countTotalNode;
        onlineNodes = countOnlineNode;
        myNodes = idsOf[addr].length;
    }

    function getIds(address addr, uint startPos) public view
    returns (uint length, bytes8[5] memory data) {
        bytes8[] memory myIds = idsOf[addr];
        length = uint(myIds.length);
        for(uint i = 0; i < 5 && (i+startPos) < length; i++) {
            data[i] = myIds[i+startPos];
        }
    }

    function has(bytes8 id) view public returns (bool)
    {
        return nodes[id].id1 != bytes32(0);
    }
}